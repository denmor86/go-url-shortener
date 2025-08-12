// Package storage предоставляет интефейсы и их реализацию для внутреннего хранения данных
package storage

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/denmor86/go-url-shortener/internal/config"
)

// TableRecord - модели записи в БД для таблицы с URL
type TableRecord struct {
	OriginalURL string // оригинальный URL, для которого был запрос на сокращение
	ShortURL    string // сокращенный URL, сформированный короткий URL
	UserID      string // идентификатор пользователя, UUID пользователя из запроса
	IsDeleted   bool   // признак необходимости удаления записи из хранилища, предполагается, что будет отдельный сервис который будет физически удалять записи из БД
}

// ReadStorage интерфейс для работы с чтением данных из хранилища
type ReadStorage interface {
	GetRecord(context.Context, string) (string, error)
	GetUserRecords(context.Context, string) ([]TableRecord, error)
	Ping(ctx context.Context) error
}

// WriteStorage интерфейс для работы с записью данных в хранилище
type WriteStorage interface {
	AddRecord(context.Context, TableRecord) error
	AddRecords(context.Context, []TableRecord) error
	DeleteURLs(context.Context, string, []string) error
	Close() error
}

// IStorage полный интерфейс для работы с хранилищем данных
type IStorage interface {
	ReadStorage
	WriteStorage
}

// NewStorage создание интерферса хранилища данных (поддерживает хранение в БД, оперативной памяти и текстовом файле)
func NewStorage(cfg *config.Config) IStorage {

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
