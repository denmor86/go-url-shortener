package logger

import (
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

var (
	once     sync.Once
	instance *zap.SugaredLogger = nil
)

// Initialize - инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string) error {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	logger, err := cfg.Build()
	if err != nil {
		return err
	}
	// устанавливаем синглтон
	instance = logger.Sugar()
	return nil
}

// Get - метод получения объекта логгера из синглтона
func Get() *zap.SugaredLogger {
	if instance == nil {
		panic("logger not initialized, call Initialize()")
	}
	return instance
}

// Sync - метод синхронизации буфферов
func Sync() error {
	if instance != nil {
		return instance.Sync()
	}
	return nil
}

// Debug — обертка над методом логирования уровня Debug
func Debug(args ...interface{}) {
	Get().Debugln(args...)
}

// Info — обертка над методом логирования уровня Info
func Info(args ...interface{}) {
	Get().Infoln(args...)
}

// Warn — обертка над методом логирования уровня Warn
func Warn(args ...interface{}) {
	Get().Warnln(args...)
}

// Error — обертка над методом логирования уровня Error
func Error(args ...interface{}) {
	Get().Errorln(args...)
}

// Panic — обертка над методом логирования уровня Panic
func Panic(args ...interface{}) {
	Get().Panicln(args...)
}

// RequestLogger — middleware-логер для входящих HTTP-запросов.
func RequestLogger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}

		h(&lw, r)

		duration := time.Since(start)

		Info("got incoming HTTP request",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
}
