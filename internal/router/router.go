package router

import (
	"github.com/gerasimovpavel/shortener.git/internal/handlers"
	"github.com/gerasimovpavel/shortener.git/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// MainRouter роутер http запросов
func MainRouter() (chi.Router, error) {
	logger, err := zap.NewDevelopment()

	if err != nil {
		return nil, err
	}
	defer logger.Sync()

	// делаем регистратор SugaredLogger
	middleware.Sugar = *logger.Sugar()

	r := chi.NewRouter()
	r.Use(
		middleware.Auth,
		middleware.Logger(logger),
		middleware.Gzip,
	)
	r.Route("/", func(r chi.Router) {
		r.Route("/{shortURL}", func(r chi.Router) {
			r.Get("/", handlers.GetHandler)
		})
		r.Get("/ping", handlers.PingHadler)
		r.Post("/", handlers.PostHandler)
		r.Route("/api", func(r chi.Router) {
			r.Route("/shorten", func(r chi.Router) {
				r.Post("/", handlers.PostJSONHandler)
				r.Post("/batch", handlers.PostJSONBatchHandler)
			})
			r.Route("/user", func(r chi.Router) {
				r.Get("/urls", handlers.GetUserURLHandler)
			})

		})

	})
	return r, err
}
