package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/crypt"
	"github.com/gerasimovpavel/shortener.git/internal/deleteuserurl"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof"
	"net/url"
	"strings"
	"testing"
	"time"
)

func Test_Handlers(t *testing.T) {
	gofakeit.Seed(0)

	urls := []struct {
		CorrelationID string `json:"correlation_id,omitempty"`
		OriginalURL   string `json:"original_url,omitempty"`
		ShortURL      string `json:"short_url,omitempty"`
	}{
		{gofakeit.UUID(),
			gofakeit.URL(),
			""},
		{gofakeit.UUID(),
			gofakeit.URL(),
			""},
		{gofakeit.UUID(),
			gofakeit.URL(),
			""},
		{gofakeit.UUID(),
			gofakeit.URL(),
			""},
	}

	tests := []struct {
		name         string
		method       string
		contenttype  string
		batch        bool
		wantStatuses []int
		resp         string
		userID       string
		hfunc        http.HandlerFunc
	}{
		{"ping storage",
			http.MethodGet,
			"ping",
			false,
			[]int{http.StatusOK},
			"",
			"53be0840-8503-11ee-b9d1-0242ac120002",
			PingHandler},
		{"ping storage error",
			http.MethodGet,
			"ping",
			true,
			[]int{http.StatusInternalServerError},
			"",
			"53be0840-8503-11ee-b9d1-0242ac120002",
			PingHandler},
		{"POST text",
			http.MethodPost,
			"plain/text",
			false,
			[]int{http.StatusCreated, http.StatusConflict},
			"",
			"53be0840-8503-11ee-b9d1-0242ac120002",
			PostHandler,
		},
		{"GET text",
			http.MethodGet,
			"plain/text",
			false,
			[]int{http.StatusTemporaryRedirect, http.StatusGone},
			"",
			"53be0840-8503-11ee-b9d1-0242ac120002",
			GetHandler,
		},
		{"POST json",
			http.MethodPost,
			"application/json",
			false,
			[]int{http.StatusCreated, http.StatusConflict},
			"",
			"53be0840-8503-11ee-b9d1-0242ac120002",
			PostJSONHandler,
		},
		{"GET json",
			http.MethodGet,
			"application/json",
			false,
			[]int{http.StatusTemporaryRedirect, http.StatusGone},
			"",
			"53be0840-8503-11ee-b9d1-0242ac120002",
			GetHandler,
		},
		{"POST json BATCH",
			http.MethodPost,
			"application/json",
			true,
			[]int{http.StatusCreated, http.StatusConflict},
			"",
			"53be0840-8503-11ee-b9d1-0242ac120002",
			PostJSONBatchHandler,
		},
		{"GET json BATCH",
			http.MethodGet,
			"application/json",
			true,
			[]int{http.StatusTemporaryRedirect},
			"",
			"53be0840-8503-11ee-b9d1-0242ac120002",
			GetHandler,
		},
		{"POST json BATCH 2",
			http.MethodPost,
			"application/json",
			true,
			[]int{http.StatusCreated, http.StatusConflict},
			"",
			"53be0840-8503-11ee-b9d1-0242ac120002",
			PostJSONBatchHandler,
		},
		{"GET User URL",
			http.MethodGet,
			"application/json",
			false,
			[]int{http.StatusOK, http.StatusNoContent},
			"",
			"53be0840-8503-11ee-b9d1-0242ac120002",
			GetUserURLHandler,
		},
		{"POST json BATCH 3",
			http.MethodPost,
			"application/json",
			true,
			[]int{http.StatusCreated, http.StatusConflict},
			"",
			"53be0840-8503-11ee-b9d1-0242ac120002",
			PostJSONBatchHandler,
		},
		{"DELETE User URL",
			http.MethodDelete,
			"application/json",
			true,
			[]int{http.StatusAccepted},
			"",
			"53be0840-8503-11ee-b9d1-0242ac120002",
			DeleteUserURLHandler,
		},
	}
	config.ParseEnvFlags()

	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			storage.Stor, err = storage.NewStorage()
			if err != nil {
				panic(err)

			}
			var target, body string
			switch tt.contenttype {
			case "ping":
				{
					switch tt.method {
					case http.MethodGet:
						{
							target = "/ping"
							if tt.batch {
								storage.Stor.Close()
							}
						}
					default:
						panic("unknown method")
					}
				}
			case "plain/text":
				{
					target = "/"
					switch tt.method {
					case http.MethodGet:
						{

							if idx == 0 {
								panic("wrong test order")
							}
							target += tests[idx-1].resp
						}
					case http.MethodPost:
						{
							body = urls[0].OriginalURL
						}
					default:
						panic("unknown method")
					}
				}
			case "application/json":
				{
					switch tt.batch {
					case false:
						{
							switch tt.method {
							case http.MethodGet:
								{
									target = "/"
									if idx == 0 {
										panic("wrong test order")
									}
									if tt.name == "GET User URL" {
										target = "/api/user/urls"
									}
									if tt.name != "GET User URL" {
										resp := new(PostResponse)
										err := json.Unmarshal([]byte(tests[idx-1].resp), &resp)
										if err != nil {
											panic(err)
										}
										u, err := url.Parse(resp.Result)
										if err != nil {
											panic(err)
										}
										target += u.Path
									}

								}
							case http.MethodPost:
								{
									target = "/api/shorten"
									body = string([]byte(fmt.Sprintf(`{"url":"%s"}`, urls[0].OriginalURL)))
								}
							default:
								panic("unknown method")
							}
						}
					case true:
						{
							jsonBatch, err := json.Marshal(urls)
							if err != nil {
								panic(err)
							}

							body = string(jsonBatch)
							switch tt.method {
							case http.MethodGet:
								{
									target = "/"
									if idx == 0 {
										panic("wrong test order")
									}

									resp := []storage.URLData{}
									err = json.Unmarshal([]byte(tests[idx-1].resp), &resp)
									if err != nil {
										panic(err)
									}
									if len(resp) == 0 {
										panic(errors.New("URL list is empty"))
									}
									u, err := url.Parse(resp[0].ShortURL)
									if err != nil {
										panic(err)
									}
									target = u.Path

								}
							case http.MethodDelete:
								{

									deleteuserurl.URLDel = deleteuserurl.NewURLDeleter()
									target = "/api/user/urls"
									req := []string{}
									resp := []storage.URLData{}
									err = json.Unmarshal([]byte(tests[idx-1].resp), &resp)
									if err != nil {
										panic(err)
									}
									if len(resp) == 0 {
										panic(errors.New("URL list is empty"))
									}
									for _, data := range resp {
										u, err := url.Parse(data.ShortURL)
										if err != nil {
											panic(err)
										}

										req = append(req, strings.TrimPrefix(u.Path, "/"))

									}
									if len(req) == 0 {
										panic(errors.New("short URL list is empty"))
									}
									s, err := json.Marshal(req)
									if err != nil {
										panic(errors.New("failed to marshalling urls"))
									}
									body = string(s)
								}
							case http.MethodPost:
								{
									target = "/api/shorten/batch"
									body = string(jsonBatch)
								}
							default:
								panic("unknown method")
							}
						}
					}
				}
			default:
				{
					panic("unknown content-type")
				}
			}
			r := strings.NewReader(body)
			req := httptest.NewRequest(tt.method, target, r)
			w := httptest.NewRecorder()

			userencrypt, err := crypt.Encrypt(tt.userID)
			if err != nil {
				panic(err)
			}
			cookie := &http.Cookie{}
			cookie.Name = "UserID"
			cookie.Expires = time.Now().Add(time.Hour * 24)
			cookie.Path = "/"
			cookie.Value = userencrypt

			http.SetCookie(w, cookie)

			var res *http.Response
			tt.hfunc(w, req)
			res = w.Result()

			b, err := io.ReadAll(res.Body)
			if err != nil {
				panic(err)
			}
			tests[idx].resp = string(b)

			res.Body.Close()

			if !assert.Contains(t, tt.wantStatuses, res.StatusCode) {
				panic(fmt.Errorf("status expect %v actual %v\nbody %v", tt.wantStatuses, res.StatusCode, body))
			}
		})
	}
}
