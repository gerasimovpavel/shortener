package main

import (
	"fmt"
	"github.com/gerasimovpavel/shortener.git/config"
	flag "github.com/spf13/pflag"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// мапа для хранения ссылок
var pairs = make(map[string]string)

// findByValue Поиск ключа по значению пары
func findByValue(value string) (key string, ok bool) {
	for k, v := range pairs {
		if v == value {
			key = k
			ok = true
			return
		}
	}
	return
}

// genShort Рандом-генератор коротких ссылок
func genShort() string {
	const allowchars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const urllen = 7
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	short := make([]byte, urllen)
	for i := range short {
		short[i] = allowchars[seed.Intn(len(allowchars))]
	}
	return string(short)
}

// postHandler handler для POST запросов
func postHandler(w http.ResponseWriter, r *http.Request) {
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
	shortURL, ok := findByValue(origURL)
	if !ok {
		shortURL = genShort()
		// записываем соотношение в мапу
		pairs[shortURL] = origURL
	}

	// Создаем URL для ответа
	tempURL := fmt.Sprintf(`%s/%s`, config.Options.ShortURLHost, shortURL)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, tempURL)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	// Определям короткую ссылку из пути
	shortURL := strings.TrimPrefix(r.URL.Path, "/")
	if shortURL == "" {
		http.Error(w, "Ссылка не указана", http.StatusBadRequest)
		return
	}
	// получаем оригинальный урл из мапы пар
	origURL, ok := pairs[shortURL]
	// при ошибки возвращаем ошибку 404
	if !ok {
		http.Error(w, "Ссылка не найдена", http.StatusNotFound)
		return
	} else {
		// иначе 307 редирект на оригинальный урл
		http.Redirect(w, r, origURL, http.StatusTemporaryRedirect)
	}
}

// mainHandler Handler для первых инкрементов
func mainHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		{
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s\n\nНе могу прочитать тело запроса", err.Error()), http.StatusBadRequest)
				return
			}
			origURL := string(body)
			if origURL == "" {
				http.Error(w, "URL в теле не найден", http.StatusBadRequest)
				return
			}
			// Создаем короткую ссылку
			shortURL, ok := findByValue(origURL)
			if !ok {
				shortURL = genShort()
				// записываем соотношение в мапу
				pairs[shortURL] = origURL
			}

			// Создаем ссылку для ответа
			tempURL := fmt.Sprintf(`%s/%s`, config.Options.ShortURLHost, shortURL)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusCreated)
			io.WriteString(w, tempURL)
		}
	case http.MethodGet:
		{
			shortURL := strings.TrimPrefix(r.URL.Path, "/")
			if shortURL == "" {
				http.Error(w, "Ссылка не указана", http.StatusBadRequest)
				return
			}
			origURL, ok := pairs[shortURL]
			// получаем оригинальный урл из мапы пар
			if !ok {
				http.Error(w, "Ссылка не найдена", http.StatusNotFound)
				return
			}
			// 307 редирект на оригинальный урл
			http.Redirect(w, r, origURL, http.StatusTemporaryRedirect)
		}
	default:
		{
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Не понял")
		}

	}
}

// mainRouter роутер http запросов
func mainRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RealIP, middleware.Logger, middleware.Recoverer)
	r.Route("/", func(r chi.Router) {
		// роут для POST
		r.Post("/", postHandler) // POST /
		r.Route("/{shortURL}", func(r chi.Router) {
			// роут для GET
			r.Get("/", getHandler) // GET /{shortURL}
		})
	})
	return r
}

// main
func main() {
	var ok bool
	// ищем переменную SERVER_ADDRESS
	config.Options.Host, ok = os.LookupEnv(`SERVER_ADDRESS`)
	if !ok {
		// если не нашли, обрабатываем командную строку
		flag.StringVarP(&config.Options.Host, "a", "a", ":8080", "Адрес HTTP-сервера")
	}
	// ищем переменную BASE_URL
	config.Options.ShortURLHost, ok = os.LookupEnv(`BASE_URL`)
	if !ok {
		// если не нашли, обрабатываем командную строку
		flag.StringVarP(&config.Options.ShortURLHost, "b", "b", "http://localhost:8080", "URL короткой ссылки")
	}
	// если хотя бы одну переменную ищем в командной строке
	if !ok {
		// парсим аргументы
		flag.Parse()
	}
	// запускаем сервер
	err := http.ListenAndServe(config.Options.Host, mainRouter())
	if err != nil {
		panic(err)
	}
}
