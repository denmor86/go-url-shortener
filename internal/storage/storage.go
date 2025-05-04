package storage

import (
	"context"
	"fmt"

	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/pkg/errors"
)

type IStorage interface {
	Add(context.Context, string, string) error
	Get(context.Context, string) (string, error)
}

func NewStorage(cfg config.Config) IStorage {

	if cfg.DatabaseDSN != "" {
		if err := CheckDSN(cfg.DatabaseDSN); err != nil {
			panic(fmt.Sprintf("invalid DSN: %s", errors.Cause(err).Error()))
		}
		storage, err := NewDatabaseStorage(cfg.DatabaseDSN)
		if err != nil {
			panic(fmt.Sprintf("can't create database storage: %s ", errors.Cause(err).Error()))
		}
		if err = storage.Initialize(cfg.DatabaseDSN); err != nil {
			panic(fmt.Sprintf("can't Initialize database storage: %s ", errors.Cause(err).Error()))
		}
		defer storage.Close()
		return storage
	}
	if cfg.FileStoragePath != "" {
		storage := NewFileStorage()
		if err := storage.Initialize(cfg.FileStoragePath); err != nil {
			panic(fmt.Sprintf("can't Initialize cache file storage: %s ", errors.Cause(err).Error()))
		}
		defer storage.Close()
		return storage
	}

	return NewMemStorage()
}
