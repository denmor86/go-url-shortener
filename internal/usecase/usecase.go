package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/helpers"
	"github.com/denmor86/go-url-shortener.git/internal/logger"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
	"github.com/denmor86/go-url-shortener.git/internal/workerpool"
)

type Usecase struct {
	Config     config.Config
	Storage    storage.IStorage
	WorkerPool *workerpool.WorkerPool
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type RequestItem struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

type ResponseItem struct {
	ID  string `json:"correlation_id"`
	URL string `json:"short_url"`
}

type ResponseURL struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

// URLDeleteJob задача удаление записей.
type URLDeleteJob struct {
	Storage   storage.IStorage
	UserID    string
	ShortURLs []string
}

// Do удаляет записи пользователя.
func (j *URLDeleteJob) Do(ctx context.Context) {
	err := j.Storage.DeleteURLs(ctx, j.UserID, j.ShortURLs)
	if err != nil {
		logger.Error("error delete URLs", err.Error())
		return
	}
	logger.Info("URLs is deleted")
}

type ContextKey string

var UserIDContextKey ContextKey = "userID"

var ErrUniqueViolation = errors.New("URL already exist")
var ErrDeletedViolation = errors.New("URL is deleted")

func NewUsecase(cfg config.Config, storage storage.IStorage, workerpool *workerpool.WorkerPool) *Usecase {
	return &Usecase{Config: cfg, Storage: storage, WorkerPool: workerpool}
}

func (u *Usecase) EncondeURL(ctx context.Context, reader io.Reader, userID string) ([]byte, error) {

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error read from body: %w", err)
	}

	url := string(data)

	if len(url) == 0 {
		return nil, fmt.Errorf("URL is empty")
	}

	shortURL, err := helpers.MakeShortURL(url, u.Config.ShortURLLen)
	if err != nil {
		return nil, fmt.Errorf("error make short URL: %w", err)
	}
	err = u.Storage.AddRecord(ctx, storage.TableRecord{OriginalURL: url, ShortURL: shortURL, UserID: userID})
	// нет ошибок
	if err == nil {
		return []byte(helpers.MakeURL(u.Config.BaseURL, shortURL)), nil
	}
	var storageError *storage.UniqueViolation
	// ошибка наличия не уникального URL
	if errors.As(err, &storageError) {
		return []byte(helpers.MakeURL(u.Config.BaseURL, storageError.ShortURL)), ErrUniqueViolation
	}
	return nil, fmt.Errorf("error storage URL: %w", err)
}

func (u *Usecase) EncondeURLJson(ctx context.Context, reader io.Reader, userID string) ([]byte, error) {

	var buf bytes.Buffer
	// читаем тело запроса
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("error read from body: %w", err)
	}
	var request Request
	if err = json.Unmarshal(buf.Bytes(), &request); err != nil {
		return nil, fmt.Errorf("error unmarshal body: %w", err)
	}

	shortURL, err := u.EncondeURL(ctx, strings.NewReader(request.URL), userID)
	var responce Response
	// нет ошибок
	if err == nil {
		responce.Result = string(shortURL)
		resp, err := json.Marshal(responce)
		if err != nil {
			return nil, fmt.Errorf("error marshaling: %w", err)
		}
		return resp, nil
	}
	// ошибка наличия не уникального URL
	if errors.Is(err, ErrUniqueViolation) {
		responce.Result = string(shortURL)
		resp, err := json.Marshal(responce)
		if err != nil {
			return nil, fmt.Errorf("error marshaling: %w", err)
		}
		return resp, ErrUniqueViolation
	}
	return nil, fmt.Errorf("error encode URL: %w", err)
}

func (u *Usecase) EncondeURLJsonBatch(ctx context.Context, reader io.Reader, userID string) ([]byte, error) {

	var buf bytes.Buffer
	// читаем тело запроса
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("error read from body: %w", err)
	}
	var requestItems []RequestItem
	if err = json.Unmarshal(buf.Bytes(), &requestItems); err != nil {
		return nil, fmt.Errorf("error unmarshal body: %w", err)
	}

	if len(requestItems) == 0 {
		return nil, fmt.Errorf("empty request: %w", err)
	}

	items := make([]storage.TableRecord, 0, len(requestItems))
	responseItems := make([]ResponseItem, 0, len(requestItems))
	for _, item := range requestItems {
		if item.ID == "" || item.URL == "" {
			return nil, fmt.Errorf("invalid request item: (ID: %s, URL: %s", item.ID, item.URL)
		}
		shortURL, err := helpers.MakeShortURL(item.URL, u.Config.ShortURLLen)
		if err != nil {
			return nil, fmt.Errorf("error make short URL: %w", err)
		}
		items = append(items, storage.TableRecord{ShortURL: shortURL, OriginalURL: item.URL, UserID: userID})
		responseItems = append(responseItems, ResponseItem{ID: item.ID, URL: helpers.MakeURL(u.Config.BaseURL, shortURL)})
	}

	if err := u.Storage.AddRecords(ctx, items); err != nil {
		return nil, fmt.Errorf("error storage urls: %w", err)
	}

	resp, err := json.Marshal(responseItems)
	if err != nil {
		return nil, fmt.Errorf("error marshaling: %w", err)
	}
	return resp, nil
}

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

func (u *Usecase) PingStorage(ctx context.Context) error {
	return u.Storage.Ping(ctx)
}

func (u *Usecase) GetURLS(ctx context.Context, userID string) ([]byte, error) {
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

	resp, err := json.Marshal(responseItems)
	if err != nil {
		return nil, fmt.Errorf("error marshaling: %w", err)
	}
	return resp, nil
}

func (u *Usecase) DeleteURLS(ctx context.Context, reader io.Reader, userID string) error {
	var buf bytes.Buffer
	// читаем тело запроса
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return fmt.Errorf("error read from body: %w", err)
	}
	var shortURLS []string
	if err = json.Unmarshal(buf.Bytes(), &shortURLS); err != nil {
		return fmt.Errorf("error unmarshal body: %w", err)
	}
	// добавляем задачу на удаление
	u.WorkerPool.AddJob(&URLDeleteJob{Storage: u.Storage, UserID: userID, ShortURLs: shortURLS})

	return nil
}
