package main

import (
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
		origURL    string
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
	var shortURL string
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var target string
			if tt.origURL != "" {
				r := strings.NewReader(tt.origURL)
				switch tt.method {
				case http.MethodPost:
					{
						target = "/"
					}
				case http.MethodGet:
					{
						target = shortURL
					}
				default:
					panic("Unknown method")
				}
				req := httptest.NewRequest(tt.method, target, r)
				w := httptest.NewRecorder()
				var res *http.Response
				switch tt.method {
				case http.MethodPost:
					{
						postHandler(w, req)
						res = w.Result()

						body, err := io.ReadAll(res.Body)
						if err != nil {
							panic(err.Error())
						}
						u, err := url.Parse(string(body))
						if err != nil {
							panic(err.Error())
						}
						shortURL = u.Path
					}
				case http.MethodGet:
					{
						getHandler(w, req)
						res = w.Result()
					}
				}
				res.Body.Close()
				if res.StatusCode != tt.wantStatus {
					panic("Не смог создать короткий URL")
				}
			}
		})
	}
}
