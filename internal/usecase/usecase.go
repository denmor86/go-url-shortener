package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/denmor86/go-url-shortener.git/internal/helpers"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

func EncondeURL(ctx context.Context, baseURL string, lenShortURL int, storage storage.IStorage, reader io.Reader) ([]byte, error) {

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error read from body: %w", err)
	}

	url := string(data)

	if len(url) == 0 {
		return nil, fmt.Errorf("URL is empty")
	}

	shortURL := helpers.MakeShortURL(url, lenShortURL)
	storage.Add(ctx, url, shortURL)

	return []byte(helpers.MakeURL(baseURL, shortURL)), nil
}

func EncondeURLJson(ctx context.Context, baseURL string, lenShortURL int, storage storage.IStorage, reader io.Reader) ([]byte, error) {

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

	shortURL, err := EncondeURL(ctx, baseURL, lenShortURL, storage, strings.NewReader(request.URL))
	if err != nil {
		return nil, fmt.Errorf("error encode URL: %w", err)
	}
	var responce Response
	responce.Result = string(shortURL)
	resp, err := json.Marshal(responce)
	if err != nil {
		return nil, fmt.Errorf("error marshaling: %w", err)
	}
	return resp, nil
}

func DecodeURL(ctx context.Context, storage storage.IStorage, shortURL string) (string, error) {

	if shortURL == "" {
		return "", fmt.Errorf("URL is empty")
	}
	url, err := storage.Get(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("error read from storage: %w", err)
	}
	return url, nil
}

func PingDatabase(ctx context.Context, dsn string, timeout time.Duration) error {
	return storage.PingPostrges(ctx, dsn, timeout)
}
