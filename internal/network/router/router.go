package router

import (
	"github.com/denmor86/go-url-shortener.git/internal/network/handlers"
	"github.com/denmor86/go-url-shortener.git/internal/network/middleware"
	"github.com/denmor86/go-url-shortener.git/internal/usecase"
	"github.com/go-chi/chi/v5"
)

func HandleRouter(u *usecase.Usecase) chi.Router {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", middleware.LogHandle(
			middleware.GzipHandle(
				handlers.EncondeURL(u)))) // POST /
		r.Post("/api/shorten", middleware.LogHandle(
			middleware.GzipHandle(
				handlers.EncondeURLJson(u)))) // POST /api/shorten
		r.Post("/api/shorten/batch", middleware.LogHandle(
			middleware.GzipHandle(
				handlers.EncondeURLJsonBatch(u)))) // POST /api/shorten/batch
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", middleware.LogHandle(handlers.DecodeURL(u))) // GET /shortURL
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", middleware.LogHandle(handlers.PingStorage(u))) // GET /shortURL
		})
	})
	return r
}
