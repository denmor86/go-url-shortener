package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/denmor86/go-url-shortener.git/internal/logger"
)

type URLInfo struct {
	ID          uint   `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	UserID      string `json:"user_uuid"`
	IsDeleted   bool   `json:"is_deleted"`
}

type FileStorage struct {
	Cache  MemStorage
	File   *os.File
	Writer *bufio.Writer
	sync.RWMutex
}

func (s *FileStorage) Close() error {
	return s.File.Close()
}

func NewFileStorage() *FileStorage {
	return &FileStorage{Cache: *NewMemStorage(), File: nil, Writer: nil}
}

func (s *FileStorage) Initialize(filepath string) error {
	if s.File != nil {
		logger.Warn("File storage already initialized")
		return nil
	}

	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	s.Writer = bufio.NewWriter(file)

	// заполняем кэш данных
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		info := URLInfo{}
		value := scanner.Text()
		err := json.Unmarshal([]byte(value), &info)
		if err != nil {
			logger.Warn("Invalid cache value has read:", value)
			continue
		}
		s.Cache.AddRecord(context.Background(), TableRecord{
			OriginalURL: info.OriginalURL,
			ShortURL:    info.ShortURL,
			UserID:      info.UserID,
			IsDeleted:   info.IsDeleted})
	}
	return nil
}

func (s *FileStorage) AddRecord(ctx context.Context, record TableRecord) error {
	s.Lock()
	s.Cache.AddRecord(ctx, record)

	info := URLInfo{ID: uint(s.Cache.Size()),
		OriginalURL: record.OriginalURL,
		ShortURL:    record.ShortURL,
		UserID:      record.UserID,
		IsDeleted:   record.IsDeleted}
	data, err := json.Marshal(&info)
	if err != nil {
		logger.Warn("Can't marchal value:", err)
	}
	// записываем значение
	if _, err := s.Writer.Write(data); err != nil {
		logger.Warn("Can't write cache value:", err)
	}
	// добавляем перенос строки
	if err := s.Writer.WriteByte('\n'); err != nil {
		logger.Warn("Invalid write separator:", err)
	}
	// записываем буфер в файл
	s.Writer.Flush()
	s.Unlock()
	return nil
}
func (s *FileStorage) AddRecords(ctx context.Context, records []TableRecord) error {
	for _, rec := range records {
		if err := s.AddRecord(ctx, rec); err != nil {
			return err
		}
	}
	return nil
}
func (s *FileStorage) GetRecord(ctx context.Context, shortURL string) (string, error) {
	s.Lock()
	longURL, err := s.Cache.GetRecord(ctx, shortURL)
	s.Unlock()
	if err == nil {
		return longURL, nil
	}
	return "", fmt.Errorf("short url not found: %s", shortURL)
}

func (s *FileStorage) GetUserRecords(ctx context.Context, userID string) ([]TableRecord, error) {
	var records []TableRecord
	s.Lock()
	for _, record := range s.Cache.Urls {
		if record.UserID == userID {
			records = append(records, record)
		}
	}
	s.Unlock()
	return records, nil
}

func (s *FileStorage) DeleteURLs(ctx context.Context, userID string, shortURLS []string) error {
	s.Lock()
	for _, shortURL := range shortURLS {
		record, exist := s.Cache.Urls[shortURL]
		if exist && record.UserID == userID {
			s.Cache.Urls[shortURL] = TableRecord{
				OriginalURL: record.OriginalURL,
				ShortURL:    record.ShortURL,
				UserID:      record.UserID,
				IsDeleted:   true}
		}
	}
	s.Unlock()
	return nil
}

func (s *FileStorage) Ping(ctx context.Context) error {
	if s.File != nil {
		return nil
	}
	return fmt.Errorf("file not open")
}
