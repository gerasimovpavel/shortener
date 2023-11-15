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
		middleware.Logger(logger),
		middleware.Gzip,
	)
	r.Route("/", func(r chi.Router) {

		// роут для POST
		r.Get("/ping", handlers.PingHadler)
		r.Post("/", handlers.PostHandler) // POST /
		r.Post("/api/shorten", handlers.PostJSONHandler)
		r.Post("/api/shorten/batch", handlers.PostJSONBatchHandler)
		r.Route("/{shortURL}", func(r chi.Router) {
			// роут для GET
			r.Get("/", handlers.GetHandler) // GET /{shortURL}
		})
	})
	return r, err
}
