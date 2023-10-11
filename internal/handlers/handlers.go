package handlers

import (
	"fmt"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	urlgen "github.com/gerasimovpavel/shortener.git/internal/urlgenerator"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"net/http"
	"strings"
)

// PostHandler handler для POST запросов
func PostHandler(w http.ResponseWriter, r *http.Request) {
	// читаем тело запросв
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// при ошибке возвращаеь 400 ошибку
		http.Error(w, fmt.Sprintf("%s\n\nНе могу прочитать тело запроса", err.Error()), http.StatusBadRequest)
		return
	}
	// Длинный URL
	origURL := string(body)
	if origURL == "" {
		http.Error(w, "URL в теле не найден", http.StatusBadRequest)
		return
	}
	// Создаем короткую ссылку
	shortURL, ok := storage.FindByValue(origURL)
	if !ok {
		shortURL = urlgen.GenShort()
		// записываем соотношение в мапу
		storage.Pairs[shortURL] = origURL
	}

	// Создаем URL для ответа
	tempURL := fmt.Sprintf(`%s/%s`, config.Options.ShortURLHost, shortURL)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, tempURL)
}

// GetHandler Хендлер для получения оригинальной ссылки
func GetHandler(w http.ResponseWriter, r *http.Request) {
	// Определям короткую ссылку из пути
	shortURL := strings.TrimPrefix(r.URL.Path, "/")
	if shortURL == "" {
		http.Error(w, "Ссылка не указана", http.StatusBadRequest)
		return
	}
	// получаем оригинальный урл из мапы пар
	origURL, ok := storage.Pairs[shortURL]
	// при ошибки возвращаем ошибку 404
	if !ok {
		http.Error(w, "Ссылка не найдена", http.StatusNotFound)
		return
	} else {
		// иначе 307 редирект на оригинальный урл
		http.Redirect(w, r, origURL, http.StatusTemporaryRedirect)
	}
}

// MainRouter роутер http запросов
func MainRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RealIP, middleware.Logger, middleware.Recoverer)
	r.Route("/", func(r chi.Router) {
		// роут для POST
		r.Post("/", PostHandler) // POST /
		r.Route("/{shortURL}", func(r chi.Router) {
			// роут для GET
			r.Get("/", GetHandler) // GET /{shortURL}
		})
	})
	return r
}
