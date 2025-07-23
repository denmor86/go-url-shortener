package main

import (
	"log"

	"github.com/denmor86/go-url-shortener/internal/app"
	"github.com/denmor86/go-url-shortener/internal/config"
	"github.com/denmor86/go-url-shortener/internal/logger"
	"github.com/denmor86/go-url-shortener/internal/storage"
)

var (
	buildVersion string = "N/A" // номер версии
	buildDate    string = "N/A" // дата сборки
	buildCommit  string = "N/A" // хэш комита
)

// showBuildInfo - вывод информации о сборке
func showBuildInfo() {
	log.Println("Build version: " + buildVersion)
	log.Println("Build date: " + buildDate)
	log.Println("Build commit: " + buildCommit)
}

// функция main вызывается автоматически при запуске приложения
func main() {

	showBuildInfo()

	config := config.NewConfig()
	storage := storage.NewStorage(config)

	defer logger.Sync()
	defer storage.Close()

	a := app.App{
		Config:  config,
		Storage: storage,
	}
	a.Run()
}
