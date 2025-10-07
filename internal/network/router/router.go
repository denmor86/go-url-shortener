// Package router предоставляет методы маршрутизации API запросов, используется пакет "github.com/go-chi/chi/v5"
package router

import (
	"net"
	"strings"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/denmor86/go-url-shortener/internal/config"
	"github.com/denmor86/go-url-shortener/internal/logger"
	"github.com/denmor86/go-url-shortener/internal/network/handlers"
	"github.com/denmor86/go-url-shortener/internal/network/middleware"
	"github.com/denmor86/go-url-shortener/internal/usecase"
)

// HandleRouter - метод формирования обработки запросов из внешнего API
func HandleRouter(cfg *config.Config, use *usecase.UsecaseHTTP) chi.Router {
	auth := middleware.NewAuthorization(cfg)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Use(middleware.LogHandle)
		r.Get("/{id}", handlers.DecodeURL(use))
		r.With(middleware.GzipHandle).With(auth.CookieHandle).
			Post("/", handlers.EncodeURL(use))
		r.Route("/api", func(r chi.Router) {
			r.Use(middleware.GzipHandle)
			r.Route("/shorten", func(r chi.Router) {
				r.Use(auth.CookieHandle)
				r.Post("/", handlers.EncodeURLJson(use))
				r.Post("/batch", handlers.EncodeURLJsonBatch(use))
			})
			r.Route("/user", func(r chi.Router) {
				r.Route("/urls", func(r chi.Router) {
					r.Use(auth.AuthHandle)
					r.Get("/", handlers.GetURLs(use))
					r.Delete("/", handlers.DeleteURLs(use))
				})
			})
			if len(cfg.TrustedSubnet) != 0 {
				_, trustedSubnet, err := net.ParseCIDR(strings.TrimSpace(cfg.TrustedSubnet))
				if err != nil {
					logger.Warn(err)
				}
				trust := middleware.NewTrustNet(trustedSubnet)
				r.Route("/internal", func(r chi.Router) {
					r.Route("/stats", func(r chi.Router) {
						r.Use(trust.TrustGuard)
						r.Get("/", handlers.GetStats(use))
					})
				})
			}
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", handlers.PingStorage(use)) // GET /ping
		})
	})

	if cfg.DebugEnable {
		r.Mount("/debug", chiMiddleware.Profiler())
	}
	return r
}
