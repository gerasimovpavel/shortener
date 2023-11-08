package main

import (
	"encoding/json"
	"errors"
	"github.com/gerasimovpavel/shortener.git/internal/handlers"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Test_main Тест основных возможностей
func Test_main(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		origURL      string
		wantStatuses []int
	}{
		{"Create short url",
			http.MethodPost,
			"https://practicum.yandex.ru",
			[]int{http.StatusCreated, http.StatusConflict},
		},
		{"Get original url",
			http.MethodGet,
			"https://practicum.yandex.ru",
			[]int{http.StatusTemporaryRedirect},
		},
		{"Get short url over json",
			http.MethodPost,
			`{"url":"https://practicum.yandex.ru"}`,
			[]int{http.StatusCreated, http.StatusConflict},
		},
	}
	err := storage.NewStorage()
	if err != nil {
		panic(err)
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
						ok := json.Valid([]byte(tt.origURL))
						if ok {
							target = "/api/shorten"
						} else {
							target = "/"
						}
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
						switch {
						case target == "/":
							{
								handlers.PostHandler(w, req)
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
						default:
							{
								handlers.PostJSONHandler(w, req)
								res = w.Result()

								body, err := io.ReadAll(res.Body)
								if err != nil {
									panic(err.Error())
								}
								ok := json.Valid(body)
								if !ok {
									panic(errors.New("json in response is invalid"))
								}
								resp := new(handlers.PostResponse)
								json.Unmarshal(body, &resp)
								u, err := url.Parse(resp.Result)
								if err != nil {
									panic(err.Error())
								}
								shortURL = u.Path
							}
						}

					}
				case http.MethodGet:
					{
						handlers.GetHandler(w, req)
						res = w.Result()
					}
				}
				res.Body.Close()
				var i int64
				for _, s := range tt.wantStatuses {
					if s == res.StatusCode {
						i++
					}
				}
				if i == 0 {
					panic("Не смог создать короткий URL")
				}
			}
		})
	}
}
