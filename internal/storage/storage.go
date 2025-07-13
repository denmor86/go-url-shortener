package storage

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/denmor86/go-url-shortener.git/internal/config"
)

type TableRecord struct {
	OriginalURL string
	ShortURL    string
	UserID      string
	IsDeleted   bool
}

type IStorage interface {
	AddRecord(context.Context, TableRecord) error
	AddRecords(context.Context, []TableRecord) error
	GetRecord(context.Context, string) (string, error)
	GetUserRecords(context.Context, string) ([]TableRecord, error)
	DeleteURLs(context.Context, string, []string) error
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
