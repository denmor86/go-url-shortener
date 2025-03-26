package app

import (
	"net/http"

	"github.com/denmor86/go-url-shortener.git/internal/network/handlers"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
)

type App struct {
	Host    string
	Port    string
	Storage storage.IStorage
}

func (a *App) Run() {
	mux := http.NewServeMux()
	mux.Handle("/", handlers.EncondeURLHandler(a.Storage))
	mux.Handle("/{id}", handlers.DecodeURLHandler(a.Storage))
	err := http.ListenAndServe(a.Host+":"+a.Port, mux)
	if err != nil {
		panic(err)
	}
}
