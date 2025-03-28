package app

import (
	"net/http"

	"github.com/denmor86/go-url-shortener.git/internal/network/router"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
)

type App struct {
	Host    string
	Port    string
	Storage storage.IStorage
}

func (a *App) Run() {

	err := http.ListenAndServe(a.Host+":"+a.Port, router.HandleRouter(a.Storage))
	if err != nil {
		panic(err)
	}
}
