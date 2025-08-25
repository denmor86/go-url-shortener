// Package usecase предоставляет реализацию бизнес логики приложения
package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/denmor86/go-url-shortener/internal/config"
	"github.com/denmor86/go-url-shortener/internal/helpers"
	"github.com/denmor86/go-url-shortener/internal/logger"
	"github.com/denmor86/go-url-shortener/internal/storage"
	"github.com/denmor86/go-url-shortener/internal/workerpool"
)

// Usecase - модель основной бизнес логики
type Usecase struct {
	Config     *config.Config         // конфигурация
	Storage    storage.IStorage       // хранилище
	WorkerPool *workerpool.WorkerPool // пул потоков
}

// URLDeleteJob - модель задачи на удаление записей
type URLDeleteJob struct {
	Storage   storage.IStorage // хранилище
	UserID    string           // UUID пользователя
	ShortURLs []string         // массив  коротких URL
}

// Do - удаляет записи пользователя.
func (j *URLDeleteJob) Do(ctx context.Context) {
	err := j.Storage.DeleteURLs(ctx, j.UserID, j.ShortURLs)
	if err != nil {
		logger.Error("error delete URLs", err.Error())
		return
	}
	logger.Info("URLs is deleted")
}

// ErrUniqueViolation - пользовательская ошибка "URL уже существует"
var ErrUniqueViolation = errors.New("URL already exist")

// ErrDeletedViolation - пользовательская ошибка "URL удален"
var ErrDeletedViolation = errors.New("URL is deleted")

// NewUsecase - метод создания объекта бизнес логики
func NewUsecase(cfg *config.Config, storage storage.IStorage, workerpool *workerpool.WorkerPool) *Usecase {
	return &Usecase{Config: cfg, Storage: storage, WorkerPool: workerpool}
}

// EncodeURL - метод формирования короткой ссылки на основе URL
func (u *Usecase) EncodeURL(ctx context.Context, url string, userID string) (string, error) {

	if len(url) == 0 {
		return "", fmt.Errorf("URL is empty")
	}

	shortURL, err := helpers.MakeShortURL(url, u.Config.ShortURLLen)
	if err != nil {
		return "", fmt.Errorf("error make short URL: %w", err)
	}
	err = u.Storage.AddRecord(ctx, storage.TableRecord{OriginalURL: url, ShortURL: shortURL, UserID: userID})
	// нет ошибок
	if err == nil {
		return helpers.MakeURL(u.Config.BaseURL, shortURL), nil
	}
	var storageError *storage.UniqueViolation
	// ошибка наличия не уникального URL
	if errors.As(err, &storageError) {
		return helpers.MakeURL(u.Config.BaseURL, storageError.ShortURL), ErrUniqueViolation
	}
	return "", fmt.Errorf("error storage URL: %w", err)
}

// EncodeURLBatch - метод формирования массива коротких ссылок
func (u *Usecase) EncodeURLBatch(ctx context.Context, requestItems []RequestItem, userID string) ([]ResponseItem, error) {

	items := make([]storage.TableRecord, 0, len(requestItems))
	responseItems := make([]ResponseItem, 0, len(requestItems))
	for _, item := range requestItems {
		if item.ID == "" || item.URL == "" {
			return nil, fmt.Errorf("invalid request item: (ID: %s, URL: %s", item.ID, item.URL)
		}
		shortURL, makeError := helpers.MakeShortURL(item.URL, u.Config.ShortURLLen)
		if makeError != nil {
			return nil, fmt.Errorf("error make short URL: %w", makeError)
		}
		items = append(items, storage.TableRecord{ShortURL: shortURL, OriginalURL: item.URL, UserID: userID})
		responseItems = append(responseItems, ResponseItem{ID: item.ID, URL: helpers.MakeURL(u.Config.BaseURL, shortURL)})
	}

	if err := u.Storage.AddRecords(ctx, items); err != nil {
		return nil, fmt.Errorf("error storage urls: %w", err)
	}

	return responseItems, nil
}

// DecodeURL - метод получения оригинального URL по короткой ссылке
func (u *Usecase) DecodeURL(ctx context.Context, shortURL string) (string, error) {
	if shortURL == "" {
		return "", fmt.Errorf("URL is empty")
	}
	url, err := u.Storage.GetRecord(ctx, shortURL)
	// нет ошибок
	if err == nil {
		return url, nil
	}
	var storageError *storage.DeletedViolation
	// ошибка: URL помечен на удаление
	if errors.As(err, &storageError) {
		return "", ErrDeletedViolation
	}
	return "", fmt.Errorf("error read from storage: %w", err)
}

// PingStorage - метод определения состояния соединения с хранилищем (БД, файл, ОП)
func (u *Usecase) PingStorage(ctx context.Context) error {
	return u.Storage.Ping(ctx)
}

// GetURLs - метод получения информации об имеющихся записях URL по пользователю
func (u *Usecase) GetURLs(ctx context.Context, userID string) ([]ResponseURL, error) {
	records, err := u.Storage.GetUserRecords(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error get user records: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	responseItems := make([]ResponseURL, 0, len(records))
	for _, item := range records {
		responseItems = append(responseItems, ResponseURL{OriginalURL: item.OriginalURL, ShortURL: helpers.MakeURL(u.Config.BaseURL, item.ShortURL)})
	}
	return responseItems, nil
}

// DeleteURLs - метод запроса на удаление информации об имеющихся записях URL по пользователю
func (u *Usecase) DeleteURLs(ctx context.Context, shortURLS []string, userID string) error {
	// добавляем задачу на удаление
	u.WorkerPool.AddJob(&URLDeleteJob{Storage: u.Storage, UserID: userID, ShortURLs: shortURLS})
	return nil
}

// GetStatistic - метод получения статистики об имеющихся записях URL и пользователях
func (u *Usecase) GetStatistic(ctx context.Context) (int, int) {
	stat := u.Storage.GetStat(ctx)
	return stat.URLs, stat.Users
}
