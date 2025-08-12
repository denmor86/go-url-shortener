package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denmor86/go-url-shortener/internal/storage"
	"github.com/denmor86/go-url-shortener/internal/usecase"
)

func TestDecodeURLHandler(t *testing.T) {

	memstorage := storage.NewMemStorage()
	memstorage.AddRecord(context.Background(), storage.TableRecord{OriginalURL: "https://practicum.yandex.ru/", ShortURL: "12345678"})
	memstorage.AddRecord(context.Background(), storage.TableRecord{OriginalURL: "https://google.com", ShortURL: "iFBc_bhG"})

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
			assert.Equal(t, tt.want.URL, string(body))
		})
	}
}
