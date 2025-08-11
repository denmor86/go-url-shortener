// Package app предоставляет реализацию инициализацию приложение
// Включает инициализацию конфига и логгера, создание сервера, запуск воркера.
package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/denmor86/go-url-shortener/internal/config"
	"github.com/denmor86/go-url-shortener/internal/helpers"
	"github.com/denmor86/go-url-shortener/internal/logger"
	"github.com/denmor86/go-url-shortener/internal/network/router"
	"github.com/denmor86/go-url-shortener/internal/storage"
	"github.com/denmor86/go-url-shortener/internal/usecase"
	"github.com/denmor86/go-url-shortener/internal/workerpool"
)

// App - модель данных приложения
type App struct {
	Config  *config.Config
	Storage storage.IStorage
}

func startServer(server *http.Server, https bool) error {
	if https {
		return server.ListenAndServeTLS("", "")
	}
	return server.ListenAndServe()
}

func createServer(listenAddr string, use *usecase.Usecase) *http.Server {
	// Генерируем самоподписанный сертификат
	cert, key, err := helpers.GenerateSelfSignedCert()
	if err != nil {
		log.Fatalf("Ошибка генерации сертификата: %v", err)
	}

	// Создаем TLS конфигурацию
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{cert},
				PrivateKey:  key,
			},
		},
	}

	return &http.Server{
		Addr:      listenAddr,
		Handler:   router.HandleRouter(use),
		TLSConfig: tlsConfig,
	}
}

// Run - метод иницилизации приложения и запуска сервера обработки сообщений
func (a *App) Run() {
	if err := logger.Initialize(a.Config.LogLevel); err != nil {
		panic(fmt.Sprintf("can't initialize logger: %s ", errors.Cause(err).Error()))
	}

	logger.Info(
		"Starting server config:", a.Config,
	)

	workerpool := workerpool.NewWorkerPool(runtime.NumCPU())
	use := usecase.NewUsecase(a.Config, a.Storage, workerpool)

	workerpool.Run()
	defer func() {
		workerpool.Close()
		logger.Info("Close worker pool...")
		workerpool.Wait()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	server := createServer(a.Config.ListenAddr, use)

	go func() {

		if err := startServer(server, a.Config.HTTPSEnabled); err != nil && err != http.ErrServerClosed {
			logger.Error("error listen server", err.Error())
		}
	}()

	<-stop
	logger.Info("Shutdown server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("error shutdown server", err.Error())
	}
	logger.Info("Server stopped")
}
