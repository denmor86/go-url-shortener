package storage

import (
	"fmt"

	"github.com/denmor86/go-url-shortener.git/internal/config"
)

type IStorage interface {
	Add(string, string) error
	Get(string) (string, error)
}

func NewStorage(cfg config.Config) IStorage {

	if cfg.DatabaseDSN != "" {
		if err := CheckDSN(cfg.DatabaseDSN); err != nil {
			panic(fmt.Sprintf("invalid DSN: %s", err.Error()))
		}
		// TODO
	}
	if cfg.FileStoragePath != "" {
		storage := NewFileStorage()
		if err := storage.Initialize(cfg.FileStoragePath); err != nil {
			panic(fmt.Sprintf("can't Initialize cache file: %s ", err.Error()))
		}
		defer storage.Close()
		return storage
	} else {
		return NewMemStorage()
	}
}
