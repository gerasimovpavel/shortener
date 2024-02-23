package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/deleteuserurl"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"github.com/gerasimovpavel/shortener.git/pkg/crypt"
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

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

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
		hfunc          http.HandlerFunc
		contenttype    string
		body           string
		erroBodyReader bool
		method         string
		name           string
		resp           string
		userID         string
		wantStatuses   []int
		batch          bool
	}{
		{
			batch:        false,
			contenttype:  "ping",
			hfunc:        PingHandler,
			method:       http.MethodGet,
			name:         "ping storage",
			resp:         "",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusOK},
		},
		{
			batch:        true,
			contenttype:  "ping",
			hfunc:        PingHandler,
			method:       http.MethodGet,
			name:         "ping storage error",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusInternalServerError},
		},
		{
			contenttype:  "plain/text",
			hfunc:        PostHandler,
			method:       http.MethodPost,
			name:         "POST text",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusCreated, http.StatusConflict},
		},
		{
			contenttype:  "plain/text",
			hfunc:        GetHandler,
			method:       http.MethodGet,
			name:         "GET text",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusTemporaryRedirect, http.StatusGone},
		},
		{
			contenttype:  "application/json",
			body:         `{url":""}`,
			hfunc:        PostJSONHandler,
			method:       http.MethodPost,
			name:         "POST json",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusBadRequest},
		},
		{
			contenttype:  "application/json",
			body:         `{"url":""}`,
			hfunc:        PostJSONHandler,
			method:       http.MethodPost,
			name:         "POST json",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusBadRequest},
		},
		{
			contenttype:    "application/json",
			erroBodyReader: true,
			hfunc:          PostJSONHandler,
			method:         http.MethodPost,
			name:           "POST json",
			userID:         "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses:   []int{http.StatusBadRequest},
		},
		{
			contenttype:  "application/json",
			hfunc:        PostJSONHandler,
			method:       http.MethodPost,
			name:         "POST json",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusCreated, http.StatusConflict},
		},
		{
			contenttype:  "application/json",
			hfunc:        GetHandler,
			method:       http.MethodGet,
			name:         "GET json",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusTemporaryRedirect, http.StatusGone},
		},
		{
			batch:        true,
			contenttype:  "application/json",
			hfunc:        PostJSONBatchHandler,
			method:       http.MethodPost,
			name:         "POST json BATCH",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusCreated, http.StatusConflict},
		},
		{
			batch:        true,
			contenttype:  "application/json",
			hfunc:        GetHandler,
			method:       http.MethodGet,
			name:         "GET json BATCH",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusTemporaryRedirect},
		},
		{
			batch:        true,
			body:         `{"url:""}`,
			contenttype:  "application/json",
			hfunc:        PostJSONBatchHandler,
			method:       http.MethodPost,
			name:         "POST json BATCH 2",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusBadRequest},
		},
		{
			batch:          true,
			erroBodyReader: true,
			contenttype:    "application/json",
			hfunc:          PostJSONBatchHandler,
			method:         http.MethodPost,
			name:           "POST json BATCH 2",
			userID:         "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses:   []int{http.StatusBadRequest},
		},
		{
			batch:        true,
			contenttype:  "application/json",
			hfunc:        PostJSONBatchHandler,
			method:       http.MethodPost,
			name:         "POST json BATCH 2",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusCreated, http.StatusConflict},
		},

		{
			contenttype:  "application/json",
			hfunc:        GetUserURLHandler,
			method:       http.MethodGet,
			name:         "GET User URL",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusOK, http.StatusNoContent},
		},
		{
			batch:        true,
			contenttype:  "application/json",
			hfunc:        PostJSONBatchHandler,
			method:       http.MethodPost,
			name:         "POST json BATCH 3",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusCreated, http.StatusConflict},
		},
		{
			batch:        true,
			contenttype:  "application/json",
			hfunc:        DeleteUserURLHandler,
			method:       http.MethodDelete,
			name:         "DELETE User URL",
			userID:       "53be0840-8503-11ee-b9d1-0242ac120002",
			wantStatuses: []int{http.StatusAccepted},
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
							if tt.body != "" {
								body = tt.body
							}
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
										err = json.Unmarshal([]byte(tests[idx-1].resp), &resp)
										if err != nil {
											panic(err)
										}
										var u *url.URL
										u, err = url.Parse(resp.Result)
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
									if tt.body != "" {
										body = tt.body
									}
								}
							default:
								panic("unknown method")
							}
						}
					case true:
						{
							var jsonBatch []byte
							jsonBatch, err = json.Marshal(urls)
							if err != nil {
								panic(err)
							}

							body = string(jsonBatch)
							if tt.body != "" {
								body = tt.body
							}
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
									var u *url.URL
									u, err = url.Parse(resp[0].ShortURL)
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
										var u *url.URL
										u, err = url.Parse(data.ShortURL)
										if err != nil {
											panic(err)
										}

										req = append(req, strings.TrimPrefix(u.Path, "/"))

									}
									if len(req) == 0 {
										panic(errors.New("short URL list is empty"))
									}
									var s []byte
									s, err = json.Marshal(req)
									if err != nil {
										panic(errors.New("failed to marshalling urls"))
									}
									body = string(s)
									if tt.body != "" {
										body = tt.body
									}
								}
							case http.MethodPost:
								{
									target = "/api/shorten/batch"
									body = string(jsonBatch)
									if tt.body != "" {
										body = tt.body
									}
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
			if tt.erroBodyReader {
				req = httptest.NewRequest(tt.method, target, errReader(0))
			}
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
