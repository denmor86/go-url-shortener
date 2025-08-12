package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/denmor86/go-url-shortener/internal/config"
	"github.com/denmor86/go-url-shortener/internal/logger"
	"github.com/denmor86/go-url-shortener/internal/storage"
	"github.com/denmor86/go-url-shortener/internal/usecase"
	"github.com/denmor86/go-url-shortener/internal/workerpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteURLS(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		response    string
	}

	tests := []struct {
		name   string
		userID string
		body   []string
		want   want
	}{
		{
			name:   "Successful",
			userID: "mda",
			body:   []string{"12345678", "iFBc_bhG"},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusAccepted,
				response:    "",
			},
		},
		{
			name:   "Empty user ID",
			userID: "",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				response:    "Undefined user\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготавливаем запрос
			bodyBytes, _ := json.Marshal(tt.body)
			request := httptest.NewRequest(
				http.MethodDelete,
				"/api/user/urls",
				bytes.NewBuffer(bodyBytes),
			)

			// Добавляем userID в контекст
			if len(tt.userID) != 0 {
				ctx := context.WithValue(request.Context(), usecase.UserIDContextKey, tt.userID)
				request = request.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			// Инициализируем usecase с нашим хранилищем
			u := newTestUsecase()

			// Вызываем обработчик
			handler := DeleteURLS(u)
			handler(w, request)

			result := w.Result()
			defer result.Body.Close()

			// Проверяем статус код
			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			// Проверяем Content-Type
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			// Проверяем тело ответа (если нужно)
			if tt.want.response != "" {
				body, err := io.ReadAll(result.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response, string(body))
			}
		})
	}
}

func newTestUsecase() *usecase.Usecase {
	// Конфигурация
	cfg := config.NewDefaultConfig()
	// Инициализация логгера
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		logger.Panic(err)
	}
	defer logger.Sync()
	// Хранение в памяти
	store := storage.NewMemStorage()
	// Тестовые записи
	store.AddRecord(context.Background(), storage.TableRecord{OriginalURL: "https://practicum.yandex.ru/", ShortURL: "12345678", UserID: "mda"})
	store.AddRecord(context.Background(), storage.TableRecord{OriginalURL: "https://google.com", ShortURL: "iFBc_bhG", UserID: "mda"})
	// Воркер для обработки удаления
	worker := workerpool.NewWorkerPool(runtime.NumCPU())
	worker.Run()

	return usecase.NewUsecase(cfg, store, worker)
}
