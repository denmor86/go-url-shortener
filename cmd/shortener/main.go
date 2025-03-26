package main

import (
	"github.com/denmor86/go-url-shortener.git/internal/app"
	"github.com/denmor86/go-url-shortener.git/internal/storage/memory"
)

// функция main вызывается автоматически при запуске приложения
func main() {
	a := app.App{
		Host:    "localhost",
		Port:    "8080",
		Storage: memory.NewMemStorage(),
	}
	a.Run()
}
