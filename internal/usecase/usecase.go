// Package usecase предоставляет реализацию бизнес логики приложения
package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/denmor86/go-url-shortener/internal/config"
	"github.com/denmor86/go-url-shortener/internal/helpers"
	"github.com/denmor86/go-url-shortener/internal/logger"
	"github.com/denmor86/go-url-shortener/internal/storage"
	"github.com/denmor86/go-url-shortener/internal/workerpool"
)

// Usecase - модель основной бизнес логики
type Usecase struct {
	Config     config.Config          // конфигурация
	Storage    storage.IStorage       // хранилище
	WorkerPool *workerpool.WorkerPool // пул потоков
}

// Request - модель запроса на формирование короткой ссылки
type Request struct {
	URL string `json:"url"` // оригинальный URL
}

// Response - модель ответа на запрос формирования короткой ссылки
type Response struct {
	Result string `json:"result"` // короткий URL
}

// RequestItem - модель запроса на формирование массива коротких ссылок
type RequestItem struct {
	ID  string `json:"correlation_id"` // UUID ссылки
	URL string `json:"original_url"`   // оригинальный URL
}

// ResponseItem - модель ответа на запрос формирования массива коротких ссылок
type ResponseItem struct {
	ID  string `json:"correlation_id"` // UUID ссылки
	URL string `json:"short_url"`      // короткий URL
}

// ResponseURL - модель ответа на запрос массива существующих у пользователя ссылок
type ResponseURL struct {
	OriginalURL string `json:"original_url"` // оригинальный URL
	ShortURL    string `json:"short_url"`    // короткий URL
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

// ContextKey - тип ключа в передаваемом контексте
type ContextKey string

// UserIDContextKey - имя ключа пользователя в передаваемом контексте
var UserIDContextKey ContextKey = "userID"

// ErrUniqueViolation - пользовательская ошибка "URL уже существует"
var ErrUniqueViolation = errors.New("URL already exist")

// ErrDeletedViolation - пользовательская ошибка "URL удален"
var ErrDeletedViolation = errors.New("URL is deleted")

// NewUsecase - метод создания объекта бизнес логики
func NewUsecase(cfg config.Config, storage storage.IStorage, workerpool *workerpool.WorkerPool) *Usecase {
	return &Usecase{Config: cfg, Storage: storage, WorkerPool: workerpool}
}

// EncodeURL - метод формирования короткой ссылки на основе тела запроса в текстовом формате
func (u *Usecase) EncodeURL(ctx context.Context, reader io.Reader, userID string) ([]byte, error) {

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

// EncodeURLJson - метод формирования короткой ссылки на основе тела запроса в JSON формате
func (u *Usecase) EncodeURLJson(ctx context.Context, reader io.Reader, userID string) ([]byte, error) {

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

	shortURL, encodeErr := u.EncodeURL(ctx, strings.NewReader(request.URL), userID)
	var responce Response
	// нет ошибок
	if encodeErr == nil {
		responce.Result = string(shortURL)
		resp, err := json.Marshal(responce)
		if err != nil {
			return nil, fmt.Errorf("error marshaling: %w", err)
		}
		return resp, nil
	}
	// ошибка наличия не уникального URL
	if errors.Is(encodeErr, ErrUniqueViolation) {
		responce.Result = string(shortURL)
		resp, err := json.Marshal(responce)
		if err != nil {
			return nil, fmt.Errorf("error marshaling: %w", err)
		}
		return resp, ErrUniqueViolation
	}
	return nil, fmt.Errorf("error encode URL: %w", encodeErr)
}

// EncodeURLJsonBatch - метод формирования массива коротких ссылок на основе тела запроса в JSON формате
func (u *Usecase) EncodeURLJsonBatch(ctx context.Context, reader io.Reader, userID string) ([]byte, error) {

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
		shortURL, makeError := helpers.MakeShortURL(item.URL, u.Config.ShortURLLen)
		if makeError != nil {
			return nil, fmt.Errorf("error make short URL: %w", makeError)
		}
		items = append(items, storage.TableRecord{ShortURL: shortURL, OriginalURL: item.URL, UserID: userID})
		responseItems = append(responseItems, ResponseItem{ID: item.ID, URL: helpers.MakeURL(u.Config.BaseURL, shortURL)})
	}

	if err = u.Storage.AddRecords(ctx, items); err != nil {
		return nil, fmt.Errorf("error storage urls: %w", err)
	}

	resp, err := json.Marshal(responseItems)
	if err != nil {
		return nil, fmt.Errorf("error marshaling: %w", err)
	}
	return resp, nil
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

// GetURLS - метод получения информации об имеющихся записях URL по пользователю
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

// DeleteURLS - метод запроса на удаление информации об имеющихся записях URL по пользователю
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
