package memory

import (
	"fmt"
	"sync"
)

type MemStorage struct {
	urlsMap map[string]string
	sync.RWMutex
}

func NewMemStorage() *MemStorage {
	var s MemStorage
	s.urlsMap = make(map[string]string)
	return &s
}

func (s *MemStorage) Save(urlBase string, urlShort string) {
	s.Lock()
	s.urlsMap[urlShort] = urlBase
	s.Unlock()
}

func (s *MemStorage) Load(urlShort string) (string, error) {
	s.Lock()
	urlBase, exist := s.urlsMap[urlShort]
	s.Unlock()
	if exist {
		return urlBase, nil
	}
	return "", fmt.Errorf("short url not found: %s", urlShort)
}
