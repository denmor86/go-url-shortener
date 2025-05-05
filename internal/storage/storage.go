package storage

import (
	"context"
	"fmt"

	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/pkg/errors"
)

type TableItem struct {
	OriginalURL string
	ShortURL    string
}

type IStorage interface {
	Add(context.Context, string, string) error
	AddMultiple(context.Context, []TableItem) error
	Get(context.Context, string) (string, error)
	Ping(ctx context.Context) error
	Close() error
}

func NewStorage(cfg config.Config) IStorage {

	if cfg.DatabaseDSN != "" {
		storage, err := NewDatabaseStorage(cfg.DatabaseDSN)
		if err != nil {
			panic(fmt.Sprintf("can't create database storage: %s ", errors.Cause(err).Error()))
		}
		if err = storage.Initialize(); err != nil {
			panic(fmt.Sprintf("can't initialize database storage: %s ", errors.Cause(err).Error()))
		}
		return storage
	}
	if cfg.FileStoragePath != "" {
		storage := NewFileStorage()
		if err := storage.Initialize(cfg.FileStoragePath); err != nil {
			panic(fmt.Sprintf("can't initialize cache file storage: %s ", errors.Cause(err).Error()))
		}
		return storage
	}

	return NewMemStorage()
}
