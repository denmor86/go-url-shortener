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

func (s *MemStorage) Add(ctx context.Context, longURL string, shortURL string) error {
	s.Lock()
	s.Urls[shortURL] = longURL
	s.Unlock()
	return nil
}

func (s *MemStorage) Get(ctx context.Context, shortURL string) (string, error) {
	s.Lock()
	longURL, exist := s.Urls[shortURL]
	s.Unlock()
	if exist {
		return longURL, nil
	}
	return "", fmt.Errorf("short url not found: %s", shortURL)
}

func (s *MemStorage) Size() int {
	return len(s.Urls)
}
