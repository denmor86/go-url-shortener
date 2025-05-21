package storage

import (
	"context"
	"fmt"
	"sync"
)

type MemStorage struct {
	Urls map[string]TableRecord
	sync.RWMutex
}

func NewMemStorage() *MemStorage {
	var s MemStorage
	s.Urls = make(map[string]TableRecord)
	return &s
}

func (s *MemStorage) Close() error {
	return nil
}

func (s *MemStorage) AddRecord(ctx context.Context, record TableRecord) error {
	s.Lock()
	s.Urls[record.ShortURL] = record
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
	record, exist := s.Urls[shortURL]
	s.Unlock()
	if exist {
		return record.OriginalURL, nil
	}
	return "", fmt.Errorf("short url not found: %s", shortURL)
}

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

func (s *MemStorage) Size() int {
	return len(s.Urls)
}

func (s *MemStorage) Ping(ctx context.Context) error {
	return nil
}
