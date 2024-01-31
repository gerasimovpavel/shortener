// Пакет для обработки хендлеров
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/deleteuserurl"
	"github.com/gerasimovpavel/shortener.git/internal/middleware"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
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

// PingHandler Хендлер для проверки работоспособности сервера
func PingHandler(w http.ResponseWriter, r *http.Request) {
	err := storage.Stor.Ping()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK")
}

// PostJSONBatchHandler Пакетное сохранение ссылок
func PostJSONBatchHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s\n\nНе могу прочитать тело запроса", err.Error()), http.StatusBadRequest)
		return
	}
	var urls []*storage.URLData
	err = json.Unmarshal(body, &urls)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s\n\nне могу десериализовать тело запроса", err.Error()), http.StatusBadRequest)
		return
	}
	for _, data := range urls {
		data.UserID = middleware.UserID
	}

	err = storage.Stor.PostBatch(urls)
	if err != nil && !errors.Is(err, storage.ErrDataConflict) {
		http.Error(w, fmt.Sprintf("не могу добавить ссылки: %v", err), http.StatusInternalServerError)
		return
	}

	for _, data := range urls {
		data.UUID = ""
		data.OriginalURL = ""
		data.UserID = ""
		data.ShortURL = fmt.Sprintf(`%s/%s`, config.Options.ShortURLHost, data.ShortURL)
	}

	var status int = http.StatusCreated
	if errors.Is(err, storage.ErrDataConflict) {
		status = http.StatusConflict
	}

	body, err = json.Marshal(urls)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s\n\nНе могу сериализовать в json", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	io.WriteString(w, string(body))
}

// PostJSONHandler Одиночное сохранение ссылки из json
func PostJSONHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s\n\nНе могу прочитать тело запроса", err.Error()), http.StatusBadRequest)
		return
	}
	pr := new(PostRequest)
	err = json.Unmarshal(body, &pr)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s\n\nне могу десериализовать тело запроса", err.Error()), http.StatusBadRequest)
		return
	}
	data := storage.URLData{}
	data.OriginalURL = pr.URL
	data.UserID = middleware.UserID

	if data.OriginalURL == "" {
		http.Error(w, "URL в теле не найден", http.StatusBadRequest)
		return
	}
	err = storage.Stor.Post(&data)
	if err != nil && !errors.Is(err, storage.ErrDataConflict) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var status int = http.StatusCreated
	if errors.Is(err, storage.ErrDataConflict) {
		status = http.StatusConflict
	}

	prp := new(PostResponse)
	prp.Result = fmt.Sprintf(`%s/%s`, config.Options.ShortURLHost, data.ShortURL)

	body, err = json.Marshal(prp)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s\n\nНе могу сериализовать json", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	io.WriteString(w, string(body))
}

// PostHandler Одиночное сохранение ссылки из plain/text
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
	data.UserID = middleware.UserID

	if data.OriginalURL == "" {
		http.Error(w, "URL в теле не найден", http.StatusBadRequest)
		return
	}
	//  СОхраняем в storage
	err = storage.Stor.Post(&data)
	if err != nil && !errors.Is(err, storage.ErrDataConflict) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//меняем статус если конфликт
	var status int = http.StatusCreated
	if errors.Is(err, storage.ErrDataConflict) {
		status = http.StatusConflict
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
	// при ошибки возвращаем ошибку 500
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка чтения: %v", err), http.StatusInternalServerError)
		return
	}

	if data.DeletedFlag {
		http.Error(w, "url has been deleted", http.StatusGone)
		return
	}
	// 307 редирект на оригинальный урл
	http.Redirect(w, r, data.OriginalURL, http.StatusTemporaryRedirect)
}

// GetUserURLHandler Хендлер для получения ссылок пользователя
func GetUserURLHandler(w http.ResponseWriter, r *http.Request) {
	urls, err := storage.Stor.GetUserURL(middleware.UserID)
	for _, data := range urls {
		data.UUID = ""
		data.UserID = ""
		data.ShortURL = fmt.Sprintf(`%s/%s`, config.Options.ShortURLHost, data.ShortURL)
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка чтения: %v", err), http.StatusInternalServerError)
		return
	}
	if len(urls) == 0 {
		http.Error(w, "no content", http.StatusNoContent)
		return
	}
	body, err := json.Marshal(urls)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s\n\nНе могу сериализовать в json", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(body))
}

// DeleteUserURLHandler Хендлер для удаления ссылко пользователя
func DeleteUserURLHandler(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, fmt.Sprintf("%s\n\nНе могу прочитать тело запроса", err.Error()), http.StatusBadRequest)
		return
	}

	s := []string{}

	err = nil
	if !json.Valid(body) {
		http.Error(w, fmt.Sprintf("неверный формат входных данных: %v", string(body)), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &s)

	if err != nil {
		// при ошибке возвращаеь 400 ошибку
		http.Error(w, fmt.Sprintf("%s\n\nне могу десериализовать тело запроса", err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	io.WriteString(w, "")

	deleteuserurl.URLDel.AddURL(&s)
}
