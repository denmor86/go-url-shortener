package memory

import (
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

func (s *MemStorage) Save(longURL string, shortURL string) {
	s.Lock()
	s.Urls[shortURL] = longURL
	s.Unlock()
}

func (s *MemStorage) Load(shortURL string) (string, error) {
	s.Lock()
	longURL, exist := s.Urls[shortURL]
	s.Unlock()
	if exist {
		return longURL, nil
	}
	return "", fmt.Errorf("short url not found: %s", shortURL)
}
