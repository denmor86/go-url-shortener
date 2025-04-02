package main

import (
	"github.com/denmor86/go-url-shortener.git/internal/app"
	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/storage/memory"
)

// функция main вызывается автоматически при запуске приложения
func main() {
	a := app.App{
		Config:  *config.NewConfig(),
		Storage: memory.NewMemStorage(),
	}
	a.Run()
}
