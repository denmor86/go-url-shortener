package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
	"github.com/denmor86/go-url-shortener.git/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncondeURLHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		bodyLen     int
	}
	tests := []struct {
		name        string
		request     string
		baseURL     string
		lenShortURL int
		body        string
		storage     storage.IStorage
		want        want
	}{
		{
			name:        "Enconde test #1 (empty body)",
			request:     "/",
			baseURL:     "http://localhost:8080",
			lenShortURL: 8,
			body:        "",
			storage:     storage.NewMemStorage(),
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyLen:     13,
			},
		},
		{
			name:        "Enconde test #2 (good body)",
			request:     "/",
			baseURL:     "http://localhost:8080/",
			lenShortURL: 8,
			body:        "https://practicum.yandex.ru/",
			storage:     storage.NewMemStorage(),
			want: want{
				contentType: "text/plain",
				statusCode:  201,
				bodyLen:     len("http://localhost:8080/") + 8,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			u := &usecase.Usecase{Storage: tt.storage, Config: config.Config{BaseURL: tt.baseURL, ShortURLLen: tt.lenShortURL}}
			h := http.HandlerFunc(EncondeURL(u))
			h(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
			//так как короткая ссылка случайная, проверяем длину тела ответа
			assert.Equal(t, tt.want.bodyLen, len(body), string(body))
		})
	}
}

func TestDecodeURLHandler(t *testing.T) {

	memstorage := storage.NewMemStorage()
	memstorage.Add(context.Background(), "https://practicum.yandex.ru/", "12345678")
	memstorage.Add(context.Background(), "https://google.com", "iFBc_bhG")

	type want struct {
		contentType string
		statusCode  int
		URL         string
	}
	tests := []struct {
		name    string
		request string
		storage storage.IStorage
		want    want
	}{
		{
			name:    "Decode test #1 (good)",
			request: "/12345678",
			storage: memstorage,
			want: want{
				contentType: "",
				statusCode:  307,
				URL:         "https://practicum.yandex.ru/",
			},
		},
		{
			name:    "Decode test #2 (url not found)",
			request: "/BIwRkGiI",
			storage: memstorage,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				URL:         "error read from storage: short url not found: BIwRkGiI\n",
			},
		},
		{
			name:    "Decode test #3 (empty url)",
			request: "/",
			storage: memstorage,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				URL:         "URL is empty\n",
			},
		},
		{
			name:    "Decode test #4 (good)",
			request: "/iFBc_bhG",
			storage: memstorage,
			want: want{
				contentType: "",
				statusCode:  307,
				URL:         "https://google.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()
			u := &usecase.Usecase{Storage: tt.storage}
			h := http.HandlerFunc(DecodeURL(u))
			h(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.want.URL, string(body))
		})
	}
}

func TestEncondeJsonURLHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		bodyLen     int
	}
	tests := []struct {
		name        string
		request     string
		baseURL     string
		lenShortURL int
		body        string
		storage     storage.IStorage
		want        want
	}{
		{
			name:        "Enconde test #1 (empty body)",
			request:     "/",
			baseURL:     "http://localhost:8080",
			lenShortURL: 8,
			body:        "",
			storage:     storage.NewMemStorage(),
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyLen:     51,
			},
		},
		{
			name:        "Enconde test #2 (bad body, invalid node)",
			request:     "/",
			baseURL:     "http://localhost:8080",
			lenShortURL: 8,
			body:        "{\"test\": \"https://practicum.yandex.ru\"}",
			storage:     storage.NewMemStorage(),
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyLen:     31,
			},
		},
		{
			name:        "Enconde test #3 (bad body, xml format)",
			request:     "/",
			baseURL:     "http://localhost:8080",
			lenShortURL: 8,
			body:        "<request><url>google.com</url></request>",
			storage:     storage.NewMemStorage(),
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyLen:     75,
			},
		},
		{
			name:        "Enconde test #4 (good body)",
			request:     "/",
			baseURL:     "http://localhost:8080/",
			lenShortURL: 8,
			body:        "{\"url\": \"https://practicum.yandex.ru\"}",
			storage:     storage.NewMemStorage(),
			want: want{
				contentType: "application/json",
				statusCode:  201,
				bodyLen:     len("http://localhost:8080/") + 21,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			u := &usecase.Usecase{Storage: tt.storage, Config: config.Config{BaseURL: tt.baseURL, ShortURLLen: tt.lenShortURL}}
			h := http.HandlerFunc(EncondeURLJson(u))
			h(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
			//так как короткая ссылка случайная, проверяем длину тела ответа
			assert.Equal(t, tt.want.bodyLen, len(body), string(body))
		})
	}
}
