package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/logger"
	"github.com/denmor86/go-url-shortener.git/internal/network/router"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
	"github.com/denmor86/go-url-shortener.git/internal/usecase"
	"github.com/pkg/errors"
)

type App struct {
	Config  config.Config
	Storage storage.IStorage
}

func (a *App) Run() {
	if err := logger.Initialize(a.Config.LogLevel); err != nil {
		panic(fmt.Sprintf("can't initialize logger: %s ", errors.Cause(err).Error()))
	}

	logger.Info(
		"Starting server config:", a.Config,
	)

	server := &http.Server{
		Addr:    a.Config.ListenAddr,
		Handler: router.HandleRouter(usecase.NewUsecase(a.Config, a.Storage)),
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
