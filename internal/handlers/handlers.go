package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gerasimovpavel/shortener.git/internal/compress/gzipp"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/log"
	"github.com/gerasimovpavel/shortener.git/internal/models"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	urlgen "github.com/gerasimovpavel/shortener.git/internal/urlgenerator"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

func PostJSONHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// при ошибке возвращаеь 400 ошибку
		http.Error(w, fmt.Sprintf("%s\n\nНе могу прочитать тело запроса", err.Error()), http.StatusBadRequest)
		return
	}
	pr := new(models.PostRequest)
	json.Unmarshal(body, &pr)
	shortURL, ok := storage.FindByValue(pr.URL)
	if !ok {
		shortURL = urlgen.GenShort()
		// записываем соотношение в мапу
		storage.Pairs[shortURL] = pr.URL
	}

	prp := new(models.PostResponse)
	prp.Result = fmt.Sprintf(`%s/%s`, config.Options.ShortURLHost, shortURL)

	body, err = json.Marshal(prp)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s\n\nНе могу сериализовать json", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, string(body))
}

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
	}
	// 307 редирект на оригинальный урл
	http.Redirect(w, r, origURL, http.StatusTemporaryRedirect)

}

// MainRouter роутер http запросов
func MainRouter() chi.Router {
	logger, err := zap.NewDevelopment()

	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// делаем регистратор SugaredLogger
	log.Sugar = *logger.Sugar()

	r := chi.NewRouter()
	r.Use(
		middleware.RealIP,
		log.Logger(logger),
		gzipp.Gzip,
		middleware.Recoverer,
	)
	r.Route("/", func(r chi.Router) {
		// роут для POST
		r.Post("/", PostHandler) // POST /
		r.Post("/api/shorten", PostJSONHandler)
		r.Route("/{shortURL}", func(r chi.Router) {
			// роут для GET
			r.Get("/", GetHandler) // GET /{shortURL}
		})
	})
	return r
}
