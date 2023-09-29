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
			origUrl := string(body)
			if origUrl == "" {
				http.Error(w, "URL в теле не найден", http.StatusBadRequest)
				return
			}
			// Создаем короткую ссылку
			shortUrl, ok := findByValue(origUrl)
			if !ok {
				shortUrl = genShort()
				// записываем соотношение в мапу
				pairs[shortUrl] = origUrl
			}

			// Создаем ссылку для ответа
			tempUrl := fmt.Sprintf(`http://localhost:8080/%s`, shortUrl)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusCreated)
			io.WriteString(w, tempUrl)
		}
	case http.MethodGet:
		{
			shortUrl := strings.TrimPrefix(r.URL.Path, "/")
			if shortUrl == "" {
				http.Error(w, "Ссылка не указана", http.StatusBadRequest)
				return
			}
			origUrl, ok := pairs[shortUrl]
			// получаем оригинальный урл из мапы пар
			if !ok {
				http.Error(w, "Ссылка не найдена", http.StatusNotFound)
				return
			}
			// 307 редирект на оригинальный урл
			http.Redirect(w, r, origUrl, http.StatusTemporaryRedirect)
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
