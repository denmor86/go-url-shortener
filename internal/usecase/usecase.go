package usecase

import (
	"bytes"
	"context"
	"encoding/json"
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
	if err != nil {
		return nil, fmt.Errorf("error storage URL: %w", err)
	}

	return []byte(helpers.MakeURL(u.Config.BaseURL, shortURL)), nil
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
