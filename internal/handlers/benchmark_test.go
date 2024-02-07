package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit"
	"github.com/gerasimovpavel/shortener.git/internal/middleware"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"github.com/gerasimovpavel/shortener.git/pkg/cookies"
	"github.com/gerasimovpavel/shortener.git/pkg/crypt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var Cookie *http.Cookie

func auth(w http.ResponseWriter, r *http.Request) {
	var err error
	if Cookie == nil {
		Cookie, _ = r.Cookie("UserID")
		err = Cookie.Valid()
		if err != nil {
			Cookie, err = cookies.NewCookie(Cookie)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
	if Cookie.Value == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	middleware.UserID, err = crypt.Decrypt(Cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if Cookie.Value == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, Cookie)
}

func BenchmarkPostHandler(b *testing.B) {
	var err error

	if storage.Stor == nil {
		storage.Stor, err = storage.NewStorage()
		if err != nil {
			panic(err)

		}
	}
	b.Run("post plain text", func(b *testing.B) {
		gofakeit.Seed(0)
		for i := 0; i < b.N; i++ {
			URL := gofakeit.URL()
			r, _ := http.NewRequest("POST", "/", strings.NewReader(URL))

			w := httptest.NewRecorder()

			auth(w, r)
			handler := http.HandlerFunc(PostHandler)

			handler.ServeHTTP(w, r)
		}
	})
}

func BenchmarkPostJSONHandler(b *testing.B) {
	var err error

	if storage.Stor == nil {
		storage.Stor, err = storage.NewStorage()
		if err != nil {
			panic(err)

		}
	}
	b.Run("post json", func(b *testing.B) {
		gofakeit.Seed(0)
		for i := 0; i < b.N; i++ {
			type url struct {
				URL string `json:"url"`
			}

			u := &url{URL: gofakeit.URL()}
			URL, err := json.Marshal(u)
			if err != nil {
				panic(fmt.Errorf("url marshalling error: %v", err))
			}
			r, _ := http.NewRequest("POST", "/", strings.NewReader(string(URL)))
			r.Header.Add("Content-Type", "application/json")

			w := httptest.NewRecorder()
			auth(w, r)

			handler := http.HandlerFunc(PostJSONHandler)

			handler.ServeHTTP(w, r)
		}
	})
}

func BenchmarkPostJSONBatchHandler(b *testing.B) {
	var err error

	if storage.Stor == nil {
		storage.Stor, err = storage.NewStorage()
		if err != nil {
			panic(err)
		}
	}
	b.Run("post batch", func(b *testing.B) {
		gofakeit.Seed(0)
		for i := 0; i < b.N; i++ {
			type url struct {
				CorrelationID string `json:"correlation_id"`
				OriginalURL   string `json:"original_url"`
			}
			var urls []*url
			for i := 0; i < 11; i++ {
				urls = append(urls,
					&url{
						CorrelationID: gofakeit.UUID(),
						OriginalURL:   gofakeit.URL()})
			}

			URLS, err := json.Marshal(urls)
			if err != nil {
				panic(fmt.Errorf("urls marshalling error: %v", err))
			}
			r, _ := http.NewRequest("POST", "/", strings.NewReader(string(URLS)))

			r.Header.Add("Content-Type", "application/json")

			w := httptest.NewRecorder()
			auth(w, r)

			handler := http.HandlerFunc(PostJSONBatchHandler)

			handler.ServeHTTP(w, r)
		}
	})
}
