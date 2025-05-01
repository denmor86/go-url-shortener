package router

import (
	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/network/handlers"
	"github.com/denmor86/go-url-shortener.git/internal/network/middleware"
	"github.com/denmor86/go-url-shortener.git/internal/usecase"
	"github.com/go-chi/chi/v5"
)

func HandleRouter(config config.Config, storage usecase.IBaseStorage) chi.Router {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", middleware.LogHandle(
			middleware.GzipHandle(
				handlers.EncondeURL(config.BaseURL, config.ShortURLLen, storage)))) // POST /
		r.Post("/api/shorten", middleware.LogHandle(
			middleware.GzipHandle(
				handlers.EncondeURLJson(config.BaseURL, config.ShortURLLen, storage)))) // POST /api/shorten
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", middleware.LogHandle(handlers.DecodeURL(storage))) // GET /shortURL
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", middleware.LogHandle(handlers.PingStorage(storage))) // GET /shortURL
		})
	})
	return r
}
