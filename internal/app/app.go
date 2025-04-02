package app

import (
	"net/http"

	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/network/router"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
)

type App struct {
	Config  config.Config
	Storage storage.IStorage
}

func (a *App) Run() {

	err := http.ListenAndServe(a.Config.ListenAddr.String(), router.HandleRouter(a.Config, a.Storage))
	if err != nil {
		panic(err)
	}
}
