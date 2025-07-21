// Package storage предоставляет интефейсы и их реализацию для внутреннего хранения данных
package storage

import (
	"context"
	"fmt"
	"sync"
)

// MemStorage - хранилище данных в кэше оперативной памяти
type MemStorage struct {
	Urls         map[string]TableRecord // записи
	sync.RWMutex                        // мьютекс для синхронизации
}

// NewMemStorage - метод создания хранилища данных в кэше из оперативной памяти
func NewMemStorage() *MemStorage {
	var s MemStorage
	s.Urls = make(map[string]TableRecord)
	return &s
}

// Close - закрытие кэша (заглушка для поддержки интерфейса)
func (s *MemStorage) Close() error {
	return nil
}

// AddRecord - метод добавления записи в кэш
func (s *MemStorage) AddRecord(ctx context.Context, record TableRecord) error {
	s.Lock()
	s.Urls[record.ShortURL] = record
	s.Unlock()
	return nil
}

// AddRecords - метод добавления массива записей в кэш
func (s *MemStorage) AddRecords(ctx context.Context, records []TableRecord) error {
	for _, rec := range records {
		if err := s.AddRecord(ctx, rec); err != nil {
			return err
		}
	}
	return nil
}

// GetRecord - метод получения записи по короткой ссылке
func (s *MemStorage) GetRecord(ctx context.Context, shortURL string) (string, error) {
	s.Lock()
	record, exist := s.Urls[shortURL]
	s.Unlock()
	if exist {
		return record.OriginalURL, nil
	}
	return "", fmt.Errorf("short url not found: %s", shortURL)
}

// GetUserRecords - метод получения массива записей пользователя из кэша в оперативной памяти
func (s *MemStorage) GetUserRecords(ctx context.Context, userID string) ([]TableRecord, error) {
	var records []TableRecord
	s.Lock()
	for _, record := range s.Urls {
		if record.UserID == userID {
			records = append(records, record)
		}
	}
	s.Unlock()
	return records, nil
}

// DeleteURLs - метод отметки массива записей пользователя на удаление
func (s *MemStorage) DeleteURLs(ctx context.Context, userID string, shortURLS []string) error {
	s.Lock()
	for _, shortURL := range shortURLS {
		record, exist := s.Urls[shortURL]
		if exist && record.UserID == userID {
			s.Urls[shortURL] = TableRecord{
				OriginalURL: record.OriginalURL,
				ShortURL:    record.ShortURL,
				UserID:      record.UserID,
				IsDeleted:   true}
		}
	}
	s.Unlock()
	return nil
}

// Size - метод определения размера кэша
func (s *MemStorage) Size() int {
	return len(s.Urls)
}

// Ping - метод проверки соединения (заглушка для поддержки интерфейса)
func (s *MemStorage) Ping(ctx context.Context) error {
	return nil
}
