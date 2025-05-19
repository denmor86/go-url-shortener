package router

import (
	"github.com/denmor86/go-url-shortener.git/internal/network/handlers"
	"github.com/denmor86/go-url-shortener.git/internal/network/middleware"
	"github.com/denmor86/go-url-shortener.git/internal/usecase"
	"github.com/go-chi/chi/v5"
)

func HandleRouter(use *usecase.Usecase) chi.Router {
	auth := middleware.NewAuthorization(use.Config)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Use(middleware.LogHandle)
		r.Get("/{id}", handlers.DecodeURL(use))
		r.With(middleware.GzipHandle).With(auth.CookieHandle).
			Post("/", handlers.EncondeURL(use))
		r.Route("/api", func(r chi.Router) {
			r.Use(middleware.GzipHandle)
			r.Route("/shorten", func(r chi.Router) {
				r.Use(auth.CookieHandle)
				r.Post("/", handlers.EncondeURLJson(use))
				r.Post("/batch", handlers.EncondeURLJsonBatch(use))
			})
			r.Route("/user", func(r chi.Router) {
				r.Route("/urls", func(r chi.Router) {
					r.Use(auth.AuthHandle)
					r.Get("/", handlers.GetURLS(use))
					r.Delete("/", handlers.DeleteURLS(use))
				})
			})
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", handlers.PingStorage(use)) // GET /ping
		})
	})
	return r
}
