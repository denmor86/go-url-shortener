package main

import (
	"github.com/denmor86/go-url-shortener.git/internal/app"
	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
)

// функция main вызывается автоматически при запуске приложения
func main() {
	config := config.NewConfig()
	storage := storage.NewStorage(config)
	a := app.App{
		Config:  config,
		Storage: storage,
	}
	a.Run()
}
