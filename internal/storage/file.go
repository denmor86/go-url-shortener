package storage

import (
	"bufio"
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
	Urls   map[string]string
	file   *os.File
	writer *bufio.Writer
	sync.RWMutex
}

func (s *FileStorage) Close() error {
	return s.file.Close()
}

func NewFileStorage() *FileStorage {
	return &FileStorage{Urls: make(map[string]string), file: nil, writer: nil}
}

func (s *FileStorage) Initialize(filepath string) error {
	if s.file != nil {
		logger.Warn("File storage already initialized")
		return nil
	}

	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	s.writer = bufio.NewWriter(file)

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
		s.Urls[info.ShortURL] = info.OriginalURL
	}
	return nil
}

func (s *FileStorage) Add(originalURL string, shortURL string) error {
	s.Lock()
	s.Urls[shortURL] = originalURL

	info := URLInfo{ID: uint(len(s.Urls)), OriginalURL: originalURL, ShortURL: shortURL}
	data, err := json.Marshal(&info)
	if err != nil {
		logger.Warn("Can't marchal value:", err)
	}
	// записываем значение
	if _, err := s.writer.Write(data); err != nil {
		logger.Warn("Can't write cache value:", err)
	}
	// добавляем перенос строки
	if err := s.writer.WriteByte('\n'); err != nil {
		logger.Warn("Invalid write separator:", err)
	}
	// записываем буфер в файл
	s.writer.Flush()
	s.Unlock()
	return nil
}

func (s *FileStorage) Get(shortURL string) (string, error) {
	s.Lock()
	longURL, exist := s.Urls[shortURL]
	s.Unlock()
	if exist {
		return longURL, nil
	}
	return "", fmt.Errorf("short url not found: %s", shortURL)
}
