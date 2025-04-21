package app

import (
	"net/http"

	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/logger"
	"github.com/denmor86/go-url-shortener.git/internal/network/handlers"
	"github.com/denmor86/go-url-shortener.git/internal/network/router"
)

type IStorage interface {
	handlers.IBaseStorage
	Initialize(string) error
	Close() error
}

type App struct {
	Config  config.Config
	Storage IStorage
}

func (a *App) Run() {
	if err := logger.Initialize(a.Config.LogLevel); err != nil {
		logger.Panic(err)
	}
	defer logger.Sync()

	if err := a.Storage.Initialize(a.Config.FileStoragePath); err != nil {
		logger.Panic("Can't Initialize cache file: ", err)
	}
	defer a.Storage.Close()

	logger.Info(
		"Starting server:", a.Config.ListenAddr,
	)
	err := http.ListenAndServe(a.Config.ListenAddr, router.HandleRouter(a.Config, a.Storage))
	if err != nil {
		logger.Panic(err)
	}
}
