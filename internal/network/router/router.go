package router

import (
	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/network/handlers"
	"github.com/denmor86/go-url-shortener.git/internal/network/middleware"
	"github.com/go-chi/chi/v5"
)

func HandleRouter(config config.Config, storage handlers.IBaseStorage) chi.Router {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", middleware.LogHandle(
			middleware.GzipHandle(
				handlers.EncondeURLHandler(config.BaseURL, config.ShortURLLen, storage)))) // POST /
		r.Post("/api/shorten", middleware.LogHandle(
			middleware.GzipHandle(
				handlers.EncondeURLJsonHandler(config.BaseURL, config.ShortURLLen, storage)))) // POST /api/shorten
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", middleware.LogHandle(handlers.DecodeURLHandler(storage))) // GET /shortURL
		})
	})
	return r
}
