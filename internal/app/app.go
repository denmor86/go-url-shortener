package app

import (
	"net/http"

	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/logger"
	"github.com/denmor86/go-url-shortener.git/internal/network/router"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
)

type App struct {
	Config  config.Config
	Storage storage.IStorage
}

func (a *App) Run() {
	if err := logger.Initialize(a.Config.LogLevel); err != nil {
		logger.Panic(err)
	}
	defer logger.Sync()

	logger.Info(
		"Starting server config:", a.Config,
	)
	err := http.ListenAndServe(a.Config.ListenAddr, router.HandleRouter(a.Config, a.Storage))
	if err != nil {
		logger.Panic(err)
	}
}
