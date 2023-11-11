package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/middleware"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
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
	err := storage.Stor.Ping()

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
	var urls []*storage.URLData
	err = json.Unmarshal(body, &urls)
	if err != nil {
		// при ошибке возвращаеь 400 ошибку
		http.Error(w, fmt.Sprintf("%s\n\nне могу десериализовать тело запроса", err.Error()), http.StatusBadRequest)
		return
	}

	// записываем в хранилище
	err = storage.Stor.PostBatch(urls)
	if err != nil && !errors.Is(err, storage.ErrDataConflict) {
		middleware.Sugar.Error(fmt.Sprintf("не могу добавить ссылки: %v", err))
		http.Error(w, fmt.Sprintf("не могу добавить ссылки: %v", err), http.StatusInternalServerError)
	}
	//самому не нравится, но это из одинаковых по названию, но разных по смыслу полей short_url
	for _, data := range urls {
		data.UUID = ""
		data.OriginalURL = ""
		data.ShortURL = fmt.Sprintf(`%s/%s`, config.Options.ShortURLHost, data.ShortURL)
	}
	//меняем статус если конфликт
	var status int
	switch errors.Is(err, storage.ErrDataConflict) {
	case true:
		{
			status = http.StatusConflict

		}
	default:
		{
			status = http.StatusCreated
		}
	}

	body, err = json.Marshal(urls)

	if err != nil {
		http.Error(w, fmt.Sprintf("%s\n\nНе могу сериализовать в json", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
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
	data := storage.URLData{}
	data.OriginalURL = pr.URL

	if data.OriginalURL == "" {
		http.Error(w, "URL в теле не найден", http.StatusBadRequest)
		return
	}
	// Сохоанем в storage
	err = storage.Stor.Post(&data)
	if err != nil && !errors.Is(err, storage.ErrDataConflict) {
		middleware.Sugar.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	//меняем статус если конфликт
	var status int
	switch errors.Is(err, storage.ErrDataConflict) {
	case true:
		{
			status = http.StatusConflict

		}
	default:
		{
			status = http.StatusCreated
		}
	}

	// Создаем URL для ответа
	prp := new(PostResponse)
	prp.Result = fmt.Sprintf(`%s/%s`, config.Options.ShortURLHost, data.ShortURL)

	body, err = json.Marshal(prp)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s\n\nНе могу сериализовать json", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
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
	data := storage.URLData{}
	// Длинный URL
	data.OriginalURL = string(body)

	if data.OriginalURL == "" {
		http.Error(w, "URL в теле не найден", http.StatusBadRequest)
		return
	}
	//  СОхраняем в storage
	err = storage.Stor.Post(&data)
	if err != nil && !errors.Is(err, storage.ErrDataConflict) {
		middleware.Sugar.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	//меняем статус если конфликт
	var status int
	switch errors.Is(err, storage.ErrDataConflict) {
	case true:
		{
			status = http.StatusConflict

		}
	default:
		{
			status = http.StatusCreated
		}
	}

	// Создаем URL для ответа
	tempURL := fmt.Sprintf(`%s/%s`, config.Options.ShortURLHost, data.ShortURL)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
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
	data, err := storage.Stor.Get(shortURL)
	// при ошибки возвращаем ошибку 404
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка чтения: %v", err), http.StatusNotFound)
		return
	}
	// 307 редирект на оригинальный урл
	http.Redirect(w, r, data.OriginalURL, http.StatusTemporaryRedirect)

}

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
		r.Get("/ping", PingHadler)
		r.Post("/", PostHandler) // POST /
		r.Post("/api/shorten", PostJSONHandler)
		r.Post("/api/shorten/batch", PostJSONBatchHandler)
		r.Route("/{shortURL}", func(r chi.Router) {
			// роут для GET
			r.Get("/", GetHandler) // GET /{shortURL}
		})
	})
	return r, err
}
