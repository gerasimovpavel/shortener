package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/cookies"
	"github.com/gerasimovpavel/shortener.git/internal/crypt"
	"github.com/gerasimovpavel/shortener.git/internal/middleware"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func auth(w http.ResponseWriter, r *http.Request) {
	middleware.UserID = ""
	cookie, _ := r.Cookie("UserID")
	err := cookie.Valid()
	if err != nil {
		cookie, err = cookies.NewCookie(cookie)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	if cookie.Value == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	middleware.UserID, err = crypt.Decrypt(cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if cookie.Value == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, cookie)
}

func BenchmarkPostHandler(b *testing.B) {
	var err error
	if config.Options.DatabaseDSN == "" {
		config.ParseEnvFlags()
	}
	if storage.Stor == nil {
		storage.Stor, err = storage.NewStorage()
		if err != nil {
			panic(err)

		}
	}
	b.Run("POST plain/text", func(b *testing.B) {
		gofakeit.Seed(0)
		for i := 0; i < b.N; i++ {
			URL := gofakeit.URL()
			r, _ := http.NewRequest("POST", "/", strings.NewReader(URL))
			w := httptest.NewRecorder()

			auth(w, r)
			handler := http.HandlerFunc(PostHandler)

			b.ReportAllocs()
			b.ResetTimer()

			handler.ServeHTTP(w, r)
		}
	})
}

func BenchmarkPostJSONHandler(b *testing.B) {
	var err error
	if config.Options.DatabaseDSN == "" {
		config.ParseEnvFlags()
	}
	if storage.Stor == nil {
		storage.Stor, err = storage.NewStorage()
		if err != nil {
			panic(err)

		}
	}
	b.Run("POST application/json", func(b *testing.B) {
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

			b.ReportAllocs()
			b.ResetTimer()

			handler.ServeHTTP(w, r)
		}
	})
}

func BenchmarkPostJSONBatchHandler(b *testing.B) {
	var err error
	if config.Options.DatabaseDSN == "" {
		config.ParseEnvFlags()
	}
	if storage.Stor == nil {
		storage.Stor, err = storage.NewStorage()
		if err != nil {
			panic(err)

		}
	}
	b.Run("POST BATCH application/json", func(b *testing.B) {
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

			b.ReportAllocs()
			b.ResetTimer()

			handler.ServeHTTP(w, r)
		}
	})
}
