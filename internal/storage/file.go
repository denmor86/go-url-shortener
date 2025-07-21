package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/denmor86/go-url-shortener/internal/logger"
)

// URLInfo - структура данных в файловом кэше
type URLInfo struct {
	ID          uint   `json:"id"`           // идентификатор записи
	OriginalURL string `json:"original_url"` // оригинальный URL
	ShortURL    string `json:"short_url"`    // короткая ссылка
	UserID      string `json:"user_uuid"`    // идетификатор пользователя
	IsDeleted   bool   `json:"is_deleted"`   // признак необходимости удаления записи
}

// FileStorage - хранилище данных в файловом кэше
type FileStorage struct {
	Cache        MemStorage    // кэш в оперативной памяти
	File         *os.File      // указатель на файловый дескриптор
	Writer       *bufio.Writer // указатель на интерфейс записи данных
	sync.RWMutex               // мьютекс для синхронизации
}

// Close - метод закрытия файлового кэша
func (s *FileStorage) Close() error {
	return s.File.Close()
}

// NewDatabaseStorage - метод создания хранилища данных в файловом кэше
func NewFileStorage() *FileStorage {
	return &FileStorage{Cache: *NewMemStorage(), File: nil, Writer: nil}
}

// Initialize - метод инициализации хранилища(создание и открытие файлового кэша)
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

// AddRecord - метод добавления записи в файловый кэш
func (s *FileStorage) AddRecord(ctx context.Context, record TableRecord) error {
	s.Lock()
	defer s.Unlock()

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
	return s.Writer.Flush()
}

// AddRecords - метод добавления массива записей в файловый кэш
func (s *FileStorage) AddRecords(ctx context.Context, records []TableRecord) error {
	for _, rec := range records {
		if err := s.AddRecord(ctx, rec); err != nil {
			return err
		}
	}
	return nil
}

// GetRecord - метод получения записи по короткой ссылке
func (s *FileStorage) GetRecord(ctx context.Context, shortURL string) (string, error) {
	s.RLock()
	defer s.RUnlock()

	longURL, err := s.Cache.GetRecord(ctx, shortURL)
	if err == nil {
		return longURL, nil
	}
	return "", fmt.Errorf("short url not found: %s", shortURL)
}

// GetUserRecords - метод получения массива записей пользователя из файлового кэша
func (s *FileStorage) GetUserRecords(ctx context.Context, userID string) ([]TableRecord, error) {
	s.RLock()
	defer s.RUnlock()

	var records []TableRecord
	for _, record := range s.Cache.Urls {
		if record.UserID == userID {
			records = append(records, record)
		}
	}
	return records, nil
}

// DeleteURLs - метод отметки массива записей пользователя на удаление
func (s *FileStorage) DeleteURLs(ctx context.Context, userID string, shortURLS []string) error {
	s.Lock()
	defer s.Unlock()

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
	return nil
}

// Ping - метод проверки наличия открытого файла с кэшем данных
func (s *FileStorage) Ping(ctx context.Context) error {
	if s.File != nil {
		return nil
	}
	return fmt.Errorf("file not open")
}
