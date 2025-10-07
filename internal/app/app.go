// Package app предоставляет реализацию инициализацию приложение
// Включает инициализацию конфига и логгера, создание сервера, запуск воркера.
package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"google.golang.org/grpc"

	"github.com/denmor86/go-url-shortener/internal/config"
	"github.com/denmor86/go-url-shortener/internal/logger"
	grpcServer "github.com/denmor86/go-url-shortener/internal/server/grpc"
	httpServer "github.com/denmor86/go-url-shortener/internal/server/http"
	"github.com/denmor86/go-url-shortener/internal/storage"
	"github.com/denmor86/go-url-shortener/internal/usecase"
	"github.com/denmor86/go-url-shortener/internal/workerpool"
)

// App - модель данных приложения
type App struct {
	Config       *config.Config
	Storage      storage.IStorage
	httpServer   *http.Server
	grpcServer   *grpc.Server
	grpcListener net.Listener
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

	workerpool.Run()
	defer func() {
		workerpool.Close()
		logger.Info("Close worker pool...")
		workerpool.Wait()
	}()

	// Запускаем серверы
	if len(a.Config.ListenAddr) > 0 {
		use := usecase.NewUsecaseHTTP(a.Config, a.Storage, workerpool)
		go a.runHTTP(use)
	}
	if len(a.Config.GRPCAddr) > 0 {
		use := usecase.NewUsecaseGRPC(a.Config, a.Storage, workerpool)
		go a.runGRPC(use)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Ждем сигнал остановки
	<-stop
	logger.Info("Shutdown signal received")
	a.shutdown()
}

// runGRPC - метод запускает GRPC сервер.
func (a *App) runGRPC(use *usecase.UsecaseGRPC) {
	listen, err := net.Listen("tcp", a.Config.GRPCAddr)
	if err != nil {
		logger.Error("GRPC server starting failed", err)
		return
	}
	a.grpcListener = listen
	a.grpcServer = grpcServer.NewServer(use)

	logger.Info("Starting GRPC server on", a.Config.GRPCAddr)

	if err := a.grpcServer.Serve(listen); err != nil {
		logger.Error("GRPC server error", err.Error())
	}
}

// runHTTP - метод запускает http сервер.
func (a *App) runHTTP(use *usecase.UsecaseHTTP) {
	a.httpServer = httpServer.NewServer(a.Config, use)
	logger.Info("Starting HTTP server on", a.Config.ListenAddr)
	if err := httpServer.StartServer(a.httpServer, a.Config.HTTPSEnabled); err != nil && err != http.ErrServerClosed {
		logger.Error("Error listen server", err.Error())
	}
}

// shutdown - метод остановки запущенных серверов
func (a *App) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown для GRPC сервера
	if a.grpcServer != nil {
		logger.Info("Stopping GRPC server gracefully...")
		a.grpcServer.GracefulStop()
		logger.Info("GRPC server stopped")
	}

	// Закрытие GRPC listener
	if a.grpcListener != nil {
		if err := a.grpcListener.Close(); err != nil {
			logger.Error("GRPC listener close error", err.Error())
		}
	}

	// Shutdown для HTTP сервера
	if a.httpServer != nil {
		logger.Info("Stopping HTTP server gracefully...")
		if err := a.httpServer.Shutdown(ctx); err != nil {
			logger.Error("HTTP server shutdown error", err.Error())
		} else {
			logger.Info("HTTP server stopped")
		}
	}
}
