package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/denmor86/go-url-shortener/internal/config"
	"github.com/denmor86/go-url-shortener/internal/storage"
	"github.com/denmor86/go-url-shortener/internal/workerpool"
)

// UsecaseHTTP - модель основной бизнес логики для HTTP
type UsecaseHTTP struct {
	use *Usecase // основная логика
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

// StatisticResponse - модель ответа на запрос статистики коротких ссылок
type StatisticResponse struct {
	URLs  int `json:"urls"`  // количество сокращенных URL
	Users int `json:"users"` // количество пользователей
}

// ContextKey - тип ключа в передаваемом контексте
type ContextKey string

// UserIDContextKey - имя ключа пользователя в передаваемом контексте
var UserIDContextKey ContextKey = "userID"

// NewUsecaseHTTP - метод создания объекта бизнес логики для HTTP запросов
func NewUsecaseHTTP(cfg *config.Config, storage storage.IStorage, workerpool *workerpool.WorkerPool) *UsecaseHTTP {
	return &UsecaseHTTP{use: &Usecase{Config: cfg, Storage: storage, WorkerPool: workerpool}}
}

// EncodeURL - метод формирования короткой ссылки на основе тела запроса в текстовом формате
func (u *UsecaseHTTP) EncodeURL(ctx context.Context, reader io.Reader, userID string) ([]byte, error) {

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error read from body: %w", err)
	}

	shortURL, err := u.use.EncodeURL(ctx, string(data), userID)

	return []byte(shortURL), err
}

// EncodeURLJson - метод формирования короткой ссылки на основе тела запроса в JSON формате
func (u *UsecaseHTTP) EncodeURLJson(ctx context.Context, reader io.Reader, userID string) ([]byte, error) {

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

	shortURL, encodeErr := u.use.EncodeURL(ctx, request.URL, userID)
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
func (u *UsecaseHTTP) EncodeURLJsonBatch(ctx context.Context, reader io.Reader, userID string) ([]byte, error) {

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

	responseItems, err := u.use.EncodeURLBatch(ctx, requestItems, userID)

	if err != nil {
		return nil, err
	}

	resp, err := json.Marshal(responseItems)
	if err != nil {
		return nil, fmt.Errorf("error marshaling: %w", err)
	}
	return resp, nil
}

// DecodeURL - метод получения оригинального URL по короткой ссылке
func (u *UsecaseHTTP) DecodeURL(ctx context.Context, shortURL string) (string, error) {
	return u.use.DecodeURL(ctx, shortURL)
}

// GetURLs - метод получения информации об имеющихся записях URL по пользователю
func (u *UsecaseHTTP) GetURLs(ctx context.Context, userID string) ([]byte, error) {
	responseItems, err := u.use.GetURLs(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp, err := json.Marshal(responseItems)
	if err != nil {
		return nil, fmt.Errorf("error marshaling: %w", err)
	}
	return resp, nil
}

// DeleteURLs - метод запроса на удаление информации об имеющихся записях URL по пользователю
func (u *UsecaseHTTP) DeleteURLs(ctx context.Context, reader io.Reader, userID string) error {
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

	u.use.DeleteURLs(ctx, shortURLS, userID)

	return nil
}

// GetStatistic - метод получения статистики об имеющихся записях URL и пользователях
func (u *UsecaseHTTP) GetStatistic(ctx context.Context) ([]byte, error) {
	URLs, Users := u.use.GetStatistic(ctx)
	resp, err := json.Marshal(StatisticResponse{URLs: URLs, Users: Users})
	if err != nil {
		return nil, fmt.Errorf("error marshaling: %w", err)
	}
	return resp, nil
}

// PingStorage - метод определения состояния соединения с хранилищем (БД, файл, ОП)
func (u *UsecaseHTTP) PingStorage(ctx context.Context) error {
	return u.use.PingStorage(ctx)
}
