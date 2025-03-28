package router

import (
	"github.com/denmor86/go-url-shortener.git/internal/network/handlers"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
	"github.com/go-chi/chi/v5"
)

func HandleRouter(storage storage.IStorage) chi.Router {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.EncondeURLHandler(storage)) // POST /
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handlers.DecodeURLHandler(storage)) // GET /shortURL
		})
	})
	return r
}
