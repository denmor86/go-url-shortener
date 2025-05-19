package storage

import (
	"context"
	"fmt"
	"sync"
)

type MemStorage struct {
	Urls map[string]string
	sync.RWMutex
}

func NewMemStorage() *MemStorage {
	var s MemStorage
	s.Urls = make(map[string]string)
	return &s
}

func (s *MemStorage) Close() error {
	return nil
}

func (s *MemStorage) AddRecord(ctx context.Context, record TableRecord) error {
	s.Lock()
	s.Urls[record.ShortURL] = record.OriginalURL
	s.Unlock()
	return nil
}

func (s *MemStorage) AddRecords(ctx context.Context, records []TableRecord) error {
	for _, rec := range records {
		if err := s.AddRecord(ctx, rec); err != nil {
			return err
		}
	}
	return nil
}

func (s *MemStorage) GetRecord(ctx context.Context, shortURL string) (string, error) {
	s.Lock()
	originalURL, exist := s.Urls[shortURL]
	s.Unlock()
	if exist {
		return originalURL, nil
	}
	return "", fmt.Errorf("short url not found: %s", shortURL)
}

func (s *MemStorage) GetUserRecords(ctx context.Context, userID string) ([]TableRecord, error) {
	var records []TableRecord
	s.Lock()
	for shortURL, originalURL := range s.Urls {
		records = append(records, TableRecord{ShortURL: shortURL, OriginalURL: originalURL})
	}
	s.Unlock()
	return records, nil
}

func (s *MemStorage) Size() int {
	return len(s.Urls)
}

func (s *MemStorage) Ping(ctx context.Context) error {
	return nil
}
