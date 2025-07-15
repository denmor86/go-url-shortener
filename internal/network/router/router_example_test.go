package router

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"strings"

	"github.com/denmor86/go-url-shortener/internal/config"
	"github.com/denmor86/go-url-shortener/internal/helpers"
	"github.com/denmor86/go-url-shortener/internal/logger"
	"github.com/denmor86/go-url-shortener/internal/storage"
	"github.com/denmor86/go-url-shortener/internal/usecase"
	"github.com/denmor86/go-url-shortener/internal/workerpool"
)

// Пример для GET /{id}
func Example_decodeURL() {
	// Инициализация роутера
	r := HandleRouter(newTestUsecase())

	req := httptest.NewRequest("GET", "/iFBc_bhG", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	fmt.Println("Status:", w.Code)
	fmt.Println("Redirect to:", w.Header().Get("Location"))
	// Output: Status: 307
	// Redirect to: https://google.com
}

// Пример для POST / (кодирование URL тело запроса)
func Example_encodeURL() {
	// Инициализация роутера
	r := HandleRouter(newTestUsecase())

	// Подготовка формы
	form := url.Values{}
	form.Add("url", "https://rambler.com/")

	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	// Output:
	// Status: 201
}

// Пример для POST /api/shorten/
func Example_encondeURLJson() {
	// Инициализация роутера
	r := HandleRouter(newTestUsecase())

	// Подготовка запроса
	jsonBody := []byte(`{"url":"https://google.com"}`)
	req := httptest.NewRequest(
		"POST",
		"/api/shorten/",
		bytes.NewBuffer(jsonBody),
	)
	// установка типа контента в заголовке
	req.Header.Set("Content-Type", "application/json")

	// Формирование ответа
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)
	// Output: Status: 201

}

// Пример для POST /api/shorten/batch
func Example_encodeBatchURLs() {
	r := HandleRouter(newTestUsecase())

	batchData := `[
        {"correlation_id": "1", "original_url": "https://test.com/batch1"},
        {"correlation_id": "2", "original_url": "https://test.com/batch2"}
    ]`
	req := httptest.NewRequest("POST", "/api/shorten/batch", strings.NewReader(batchData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)
	// Output:
	// Status: 201
}

// Example_getUserURLs - Пример для GET /api/user/urls
func Example_getUserURLs() {
	r := HandleRouter(newTestUsecase())
	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	if token, err := helpers.BuildJWT("mda", []byte("secret")); err == nil {
		req.AddCookie(&http.Cookie{Name: "user-token", Value: token})
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("URLs:", w.Body.String())
	// Output:
	// Status: 200
	// URLs: [{"original_url":"https://practicum.yandex.ru/","short_url":"http://localhost:8080/12345678"},{"original_url":"https://google.com","short_url":"http://localhost:8080/iFBc_bhG"}]
}

// Пример для DELETE /api/user/urls
func Example_deleteUserURLs() {
	r := HandleRouter(newTestUsecase())
	req := httptest.NewRequest("DELETE", "/api/user/urls", strings.NewReader(`["iFBc_bhG"]`))
	req.Header.Set("Content-Type", "application/json")

	if token, err := helpers.BuildJWT("mda", []byte("secret")); err == nil {
		req.AddCookie(&http.Cookie{Name: "user-token", Value: token})
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)
	// Output:
	// Status: 202
}

// Пример для GET /ping
func Example_ping() {
	r := HandleRouter(newTestUsecase())

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)
	// Output: Status: 200
}

func newTestUsecase() *usecase.Usecase {
	// Конфигурация
	config := config.DefaultConfig()
	// Инициализация логгера
	if err := logger.Initialize(config.LogLevel); err != nil {
		logger.Panic(err)
	}
	defer logger.Sync()
	// Хранение в памяти
	store := storage.NewStorage(config)
	// Тестовые записи
	store.AddRecord(context.Background(), storage.TableRecord{OriginalURL: "https://practicum.yandex.ru/", ShortURL: "12345678", UserID: "mda"})
	store.AddRecord(context.Background(), storage.TableRecord{OriginalURL: "https://google.com", ShortURL: "iFBc_bhG", UserID: "mda"})
	// Воркер для обработки удаления
	worker := workerpool.NewWorkerPool(runtime.NumCPU())
	worker.Run()

	return usecase.NewUsecase(config, store, worker)
}
