package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/denmor86/go-url-shortener.git/internal/helpers"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

func EncondeURL(baseURL string, lenShortURL int, storage storage.IStorage, reader io.Reader) ([]byte, error) {

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Error read from body", err)
	}

	url := string(data)

	if len(url) == 0 {
		return nil, fmt.Errorf("URL is empty")
	}

	shortURL := helpers.MakeShortURL(url, lenShortURL)
	storage.Add(url, shortURL)

	return []byte(helpers.MakeURL(baseURL, shortURL)), nil
}

func EncondeURLJson(baseURL string, lenShortURL int, storage storage.IStorage, reader io.Reader) ([]byte, error) {

	var buf bytes.Buffer
	// читаем тело запроса
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Error read from body", err)
	}
	var request Request
	if err = json.Unmarshal(buf.Bytes(), &request); err != nil {
		return nil, fmt.Errorf("%s: %w", "Error unmarshal body", err)
	}

	shortURL, err := EncondeURL(baseURL, lenShortURL, storage, strings.NewReader(request.URL))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Error encode URL", err)
	}
	var responce Response
	responce.Result = string(shortURL)
	resp, err := json.Marshal(responce)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Error marshaling", err)
	}
	return resp, nil
}

func DecodeURL(storage storage.IStorage, shortURL string) (string, error) {

	if shortURL == "" {
		return "", fmt.Errorf("URL is empty")
	}
	url, err := storage.Get(shortURL)
	if err != nil {
		return "", fmt.Errorf("%s: %w", "Error read from storage", err)
	}
	return url, nil
}
