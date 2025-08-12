package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/denmor86/go-url-shortener/internal/config"
	"github.com/denmor86/go-url-shortener/internal/storage"
	"github.com/denmor86/go-url-shortener/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUserID = "c9c5cb66-dbbc-4d57-8cb9-55f58096f79b"
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
			u := &usecase.Usecase{Storage: tt.storage, Config: &config.Config{BaseURL: tt.baseURL, ShortURLLen: tt.lenShortURL}}
			h := http.HandlerFunc(EncondeURL(u))
			ctx := request.Context()
			ctx = context.WithValue(ctx, usecase.UserIDContextKey, testUserID)
			request = request.WithContext(ctx)
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
			u := &usecase.Usecase{Storage: tt.storage, Config: &config.Config{BaseURL: tt.baseURL, ShortURLLen: tt.lenShortURL}}
			h := http.HandlerFunc(EncondeURLJson(u))
			ctx := request.Context()
			ctx = context.WithValue(ctx, usecase.UserIDContextKey, testUserID)
			request = request.WithContext(ctx)
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

func TestEncondeJsonURLHandlerBatch(t *testing.T) {
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
			body:        `[{"correlation_id":"c978edb7-eb81-45b3-bcc7-e5cf9f5781cd","short_id":"http://qpabthuzw1vjfl.com"},{"correlation_id":"0a5c6e26-f875-44c8-9e09-1ceefa82235e","short_id":"http://nqea9x1nxhuinc.biz/cvn6iupy"}]`,
			storage:     storage.NewMemStorage(),
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyLen:     71,
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
			body:        `[{"correlation_id":"c978edb7-eb81-45b3-bcc7-e5cf9f5781cd","original_url":"http://qpabthuzw1vjfl.com"},{"correlation_id":"0a5c6e26-f875-44c8-9e09-1ceefa82235e","original_url":"http://nqea9x1nxhuinc.biz/cvn6iupy"}]`,
			storage:     storage.NewMemStorage(),
			want: want{
				contentType: "application/json",
				statusCode:  201,
				bodyLen:     207,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			u := &usecase.Usecase{Storage: tt.storage, Config: &config.Config{BaseURL: tt.baseURL, ShortURLLen: tt.lenShortURL}}
			h := http.HandlerFunc(EncondeURLJsonBatch(u))
			ctx := request.Context()
			ctx = context.WithValue(ctx, usecase.UserIDContextKey, testUserID)
			request = request.WithContext(ctx)
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
