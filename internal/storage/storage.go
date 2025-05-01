package storage

import (
	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/logger"
)

type IStorage interface {
	Add(string, string) error
	Get(string) (string, error)
}

func NewStorage(cfg config.Config) IStorage {

	if cfg.FileStoragePath != "" {
		storage := NewFileStorage()
		if err := storage.Initialize(cfg.FileStoragePath); err != nil {
			logger.Panic("Can't Initialize cache file: ", err)
		}
		defer storage.Close()
		return storage
	} else {
		return NewMemStorage()
	}
}
