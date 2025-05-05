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
		s.Cache.Add(context.Background(), info.OriginalURL, info.ShortURL)
	}
	return nil
}

func (s *FileStorage) Add(ctx context.Context, originalURL string, shortURL string) error {
	s.Lock()
	s.Cache.Add(ctx, originalURL, shortURL)

	info := URLInfo{ID: uint(s.Cache.Size()), OriginalURL: originalURL, ShortURL: shortURL}
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

func (s *FileStorage) Get(ctx context.Context, shortURL string) (string, error) {
	s.Lock()
	longURL, err := s.Cache.Get(ctx, shortURL)
	s.Unlock()
	if err == nil {
		return longURL, nil
	}
	return "", fmt.Errorf("short url not found: %s", shortURL)
}
func (s *FileStorage) Ping(ctx context.Context) error {
	if s.File != nil {
		return nil
	}
	return fmt.Errorf("file not open")
}
