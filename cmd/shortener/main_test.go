package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func Test_mainHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		origUrl    string
		wantStatus int
	}{
		{"Create short url",
			http.MethodPost,
			"https://practicum.yandex.ru",
			http.StatusCreated},
		{"Get original url",
			http.MethodGet,
			"https://practicum.yandex.ru",
			http.StatusTemporaryRedirect},
	}
	var shortUrl string
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var target string
			if tt.origUrl != "" {
				r := strings.NewReader(tt.origUrl)
				switch tt.method {
				case http.MethodPost:
					{
						target = "/"
					}
				case http.MethodGet:
					{
						target = fmt.Sprintf("%s", shortUrl)
					}
				default:
					panic("Unknown method")
				}
				req := httptest.NewRequest(tt.method, target, r)
				w := httptest.NewRecorder()
				mainHandler(w, req)
				res := w.Result()
				if tt.method == http.MethodPost {
					body, err := io.ReadAll(res.Body)
					if err != nil {
						panic(err.Error())
					}
					u, err := url.Parse(string(body))
					shortUrl = u.Path
				}

				if res.StatusCode != tt.wantStatus {
					panic("Не смог создать короткий URL")
				}
			}
		})
	}
}
