package router

import (
	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/logger"
	"github.com/denmor86/go-url-shortener.git/internal/network/handlers"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
	"github.com/go-chi/chi/v5"
)

func HandleRouter(config config.Config, storage storage.IStorage) chi.Router {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", logger.RequestLogger(handlers.EncondeURLHandler(config.BaseURL, config.ShortURLLen, storage)))                // POST /
		r.Post("/api/shorten", logger.RequestLogger(handlers.EncondeURLJsonHandler(config.BaseURL, config.ShortURLLen, storage))) // POST /api/shorten
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", logger.RequestLogger(handlers.DecodeURLHandler(storage))) // GET /shortURL
		})
	})
	return r
}
