package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var pairs = make(map[string]string)

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
			tempURL := fmt.Sprintf(`http://localhost:8080/%s`, shortURL)
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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", mainHandler)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
