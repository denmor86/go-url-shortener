// Package middleware предоставляет впомогательные middleware методы для поддержки сетевого взаимодействия
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/denmor86/go-url-shortener/internal/logger"
)

type (
	// responseData - структура для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// loggingResponseWriter - реализация пользовательского http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter               // оригинальный http.ResponseWriter
		responseData        *responseData // сведения об ответе
	}
)

var (
	// loggingWriterPool - пул пользовательских http.ResponseWriter
	loggingWriterPool = sync.Pool{
		New: func() any {
			return &loggingResponseWriter{}
		},
	}
)

// WriteHeader - метод пользовательской записи тела запроса
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

// WriteHeader - метод пользовательской записи заголовка запроса
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

// LogHandle — middleware-логер для входящих HTTP-запросов.
func LogHandle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		// Получаем из пула
		lw := loggingWriterPool.Get().(*loggingResponseWriter)
		lw.ResponseWriter = w
		lw.responseData = responseData

		h.ServeHTTP(lw, r)

		duration := time.Since(start)

		logger.Info("got incoming HTTP request",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)

		// Возвращаем в пул
		loggingWriterPool.Put(lw)
	})
}
