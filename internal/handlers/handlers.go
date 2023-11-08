package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/middleware"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	urlgen "github.com/gerasimovpavel/shortener.git/internal/urlgenerator"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

type PostRequest struct {
	URL string `json:"url"`
}
type PostResponse struct {
	Result string `json:"result"`
}

func PingHadler(w http.ResponseWriter, r *http.Request) {
	err := storage.Ping()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK")
}

func PostJSONBatchHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// при ошибке возвращаеь 400 ошибку
		http.Error(w, fmt.Sprintf("%s\n\nНе могу прочитать тело запроса", err.Error()), http.StatusBadRequest)
		return
	}
	var data []*storage.URLData
	err = json.Unmarshal(body, &data)
	if err != nil {
		// при ошибке возвращаеь 400 ошибку
		http.Error(w, fmt.Sprintf("%s\n\nне могу десериализовать тело запроса", err.Error()), http.StatusBadRequest)
		return
	}
	// записываем соотношение в хранилище
	err = storage.PostBatch(data)
	if err != nil {
		middleware.Sugar.Error(fmt.Sprintf("не могу добавить ссылки: %v", err))
		http.Error(w, fmt.Sprintf("не могу добавить ссылки: %v", err), http.StatusInternalServerError)
	}
	var IsConflict bool
	for _, url := range data {
		if url.IsConflict {
			IsConflict = true
		}
		if url.ShortURL == "" {
			http.Error(w, "Не все ссылки обработаны", http.StatusConflict)
			break
		}
		url.ShortURL = fmt.Sprintf("%s/%s", config.Options.ShortURLHost, strings.Trim(url.ShortURL, " "))
		url.OriginalURL = ""
		url.UUID = ""

	}
	body, err = json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s\n\nНе могу сериализовать json", err.Error()), http.StatusBadGateway)
	}

	w.Header().Set("Content-Type", "application/json")
	switch IsConflict {
	case true:
		{
			w.WriteHeader(http.StatusConflict)
		}
	default:
		{
			w.WriteHeader(http.StatusCreated)
		}
	}
	io.WriteString(w, string(body))
}

func PostJSONHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// при ошибке возвращаеь 400 ошибку
		http.Error(w, fmt.Sprintf("%s\n\nНе могу прочитать тело запроса", err.Error()), http.StatusBadRequest)
		return
	}
	pr := new(PostRequest)
	json.Unmarshal(body, &pr)
	data, err := storage.FindByOriginalURL(pr.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка поиска по оргинальной ссылке: %v", err), http.StatusInternalServerError)
	}

	switch data.ShortURL {
	case "":
		{
			data.ShortURL = urlgen.GenShort()
			data.OriginalURL = pr.URL
			err = storage.Post(data)
			if err != nil {
				http.Error(w, fmt.Sprintf("не могу добавить ссылку: %v", err), http.StatusInternalServerError)
			}

		}
	default:
		{
			data.IsConflict = true
		}
	}

	prp := new(PostResponse)
	prp.Result = fmt.Sprintf(`%s/%s`, config.Options.ShortURLHost, strings.Trim(data.ShortURL, " "))

	body, err = json.Marshal(prp)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s\n\nНе могу сериализовать json", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	switch data.IsConflict {
	case true:
		{
			w.WriteHeader(http.StatusConflict)
		}
	default:
		{
			w.WriteHeader(http.StatusCreated)
		}
	}

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
	data, err := storage.FindByOriginalURL(origURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка поиска по оргинальной ссылке: %v", err), http.StatusInternalServerError)
	}
	if data.ShortURL == "" {
		data.ShortURL = urlgen.GenShort()
		data.OriginalURL = origURL
		// записываем соотношение
		err = storage.Post(data)
		if err != nil {
			http.Error(w, fmt.Sprintf("не могу добавить ссылку: %v", err), http.StatusInternalServerError)
		}
	}

	// Создаем URL для ответа
	tempURL := fmt.Sprintf(`%s/%s`, config.Options.ShortURLHost, data.ShortURL)
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
	data, err := storage.Get(shortURL)
	// при ошибки возвращаем ошибку 404
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка чтения: %v", err), http.StatusNotFound)
		return
	}
	// 307 редирект на оригинальный урл
	http.Redirect(w, r, data.OriginalURL, http.StatusTemporaryRedirect)

}

// MainRouter роутер http запросов
func MainRouter() chi.Router {
	logger, err := zap.NewDevelopment()

	if err != nil {
		panic(err)
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
		r.Get("/ping", PingHadler)
		r.Post("/", PostHandler) // POST /
		r.Post("/api/shorten", PostJSONHandler)
		r.Post("/api/shorten/batch", PostJSONBatchHandler)
		r.Route("/{shortURL}", func(r chi.Router) {
			// роут для GET
			r.Get("/", GetHandler) // GET /{shortURL}
		})
	})
	return r
}
