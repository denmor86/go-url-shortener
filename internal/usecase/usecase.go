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
	"github.com/denmor86/go-url-shortener.git/internal/storage"
)

type Usecase struct {
	Config  config.Config
	Storage storage.IStorage
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

var ErrUniqueViolation = errors.New("URL already exist")

func NewUsecase(cfg config.Config, storage storage.IStorage) *Usecase {
	return &Usecase{Config: cfg, Storage: storage}
}

func (u *Usecase) EncondeURL(ctx context.Context, reader io.Reader) ([]byte, error) {

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error read from body: %w", err)
	}

	url := string(data)

	if len(url) == 0 {
		return nil, fmt.Errorf("URL is empty")
	}

	shortURL := helpers.MakeShortURL(url, u.Config.ShortURLLen)
	err = u.Storage.Add(ctx, url, shortURL)
	// нет ошибок
	if err == nil {
		return []byte(helpers.MakeURL(u.Config.BaseURL, shortURL)), nil
	}
	var storageError *storage.UniqueViolationError
	// ошибка наличия не уникального URL
	if errors.As(err, &storageError) {
		return []byte(helpers.MakeURL(u.Config.BaseURL, storageError.ShortURL)), ErrUniqueViolation
	}
	return nil, fmt.Errorf("error storage URL: %w", err)
}

func (u *Usecase) EncondeURLJson(ctx context.Context, reader io.Reader) ([]byte, error) {

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

	shortURL, err := u.EncondeURL(ctx, strings.NewReader(request.URL))
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

func (u *Usecase) EncondeURLJsonBatch(ctx context.Context, reader io.Reader) ([]byte, error) {

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

	items := make([]storage.TableItem, 0, len(requestItems))
	responseItems := make([]ResponseItem, 0, len(requestItems))
	for _, item := range requestItems {
		if item.ID == "" || item.URL == "" {
			return nil, fmt.Errorf("invalid request item: (ID: %s, URL: %s", item.ID, item.URL)
		}
		shortURL := helpers.MakeShortURL(item.URL, u.Config.ShortURLLen)
		items = append(items, storage.TableItem{ShortURL: shortURL, OriginalURL: item.URL})
		responseItems = append(responseItems, ResponseItem{ID: item.ID, URL: helpers.MakeURL(u.Config.BaseURL, shortURL)})
	}

	if err := u.Storage.AddMultiple(ctx, items); err != nil {
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
	url, err := u.Storage.Get(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("error read from storage: %w", err)
	}
	return url, nil
}

func (u *Usecase) PingStorage(ctx context.Context) error {
	return u.Storage.Ping(ctx)
}
