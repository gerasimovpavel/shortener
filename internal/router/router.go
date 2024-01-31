package router

import (
	"github.com/gerasimovpavel/shortener.git/internal/handlers"
	"github.com/gerasimovpavel/shortener.git/internal/logger"
	mw "github.com/gerasimovpavel/shortener.git/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// MainRouter роутер http запросов
func MainRouter() chi.Router {
	// делаем регистратор SugaredLogger
	mw.Sugar = *logger.Logger.Sugar()

	r := chi.NewRouter()
	r.Mount("/debug", middleware.Profiler())
	r.Group(func(r chi.Router) {
		r.Use(
			mw.AuthHeader,
			mw.Logger(logger.Logger),
			mw.Gzip,
		)

		r.Get("/{shortURL}", handlers.GetHandler)
		r.Get("/ping", handlers.PingHandler)
		r.Post("/", handlers.PostHandler)
		r.Route("/api", func(r chi.Router) {
			r.Route("/shorten", func(r chi.Router) {
				r.Post("/", handlers.PostJSONHandler)
				r.Post("/batch", handlers.PostJSONBatchHandler)
			})
			r.Route("/user", func(r chi.Router) {
				r.Get("/urls", handlers.GetUserURLHandler)
				r.Delete("/urls", handlers.DeleteUserURLHandler)
			})

		})
	})
	return r
}
