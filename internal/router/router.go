// Package router реализует chi роутер для маршрутизации http запросов
package router

import (
	"github.com/gerasimovpavel/shortener.git/internal/handlers"
	mw "github.com/gerasimovpavel/shortener.git/internal/middleware"
	"github.com/gerasimovpavel/shortener.git/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// MainRouter роутер http запросов
func MainRouter() chi.Router {
	// делаем регистратор SugaredLogger
	mw.Sugar = *logger.Logger.Sugar()

	r := chi.NewRouter()
	r.Mount("/debug", middleware.Profiler())
	r.Get("/{shortURL}", handlers.GetHandler)
	r.Get("/ping", handlers.PingHandler)
	r.Group(func(r chi.Router) {
		r.Use(
			mw.AutoAuthHeader,
			mw.Logger(logger.Logger),
			mw.Gzip,
		)
		r.Post("/", handlers.PostHandler)
		r.Route("/api", func(r chi.Router) {
			r.Route("/internal", func(r chi.Router) {
				r.Use(middleware.RealIP)
				r.Get("/stats", handlers.GetStatHandler)
			})
			r.Route("/shorten", func(r chi.Router) {
				r.Post("/", handlers.PostJSONHandler)
				r.Post("/batch", handlers.PostJSONBatchHandler)
			})

			r.Route("/user", func(r chi.Router) {
				r.Delete("/urls", handlers.DeleteUserURLHandler)
				r.Group(func(r chi.Router) {
					r.Use(mw.AuthHeader)
					r.Get("/urls", handlers.GetUserURLHandler)
				})

			})

		})
	})

	return r
}
