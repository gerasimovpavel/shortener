package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func Test_MainRouter(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"create router"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MainRouter()
			if err != nil {
				panic(err)
			}
		})
	}
}

func Test_Handlers(t *testing.T) {
	gofakeit.Seed(0)

	urls := []struct {
		CorrelationId string `json:"correlation_id,omitempty"`
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
		hfunc        http.HandlerFunc
	}{
		{"ping storage",
			http.MethodGet,
			"ping",
			false,
			[]int{http.StatusOK},
			"",
			PingHadler},
		{"ping storage error",
			http.MethodGet,
			"ping",
			true,
			[]int{http.StatusInternalServerError},
			"",
			PingHadler},
		{"POST text",
			http.MethodPost,
			"plain/text",
			false,
			[]int{http.StatusCreated, http.StatusConflict},
			"",
			PostHandler,
		},
		{"GET text",
			http.MethodGet,
			"plain/text",
			false,
			[]int{http.StatusTemporaryRedirect},
			"",
			GetHandler,
		},
		{"POST json",
			http.MethodPost,
			"application/json",
			false,
			[]int{http.StatusCreated, http.StatusConflict},
			"",
			PostJSONHandler,
		},
		{"GET json",
			http.MethodGet,
			"application/json",
			false,
			[]int{http.StatusTemporaryRedirect},
			"",
			GetHandler,
		},
		{"POST json BATCH",
			http.MethodPost,
			"application/json",
			true,
			[]int{http.StatusCreated, http.StatusConflict},
			"",
			PostJSONBatchHandler,
		},
		{"GET json BATCH",
			http.MethodGet,
			"application/json",
			true,
			[]int{http.StatusTemporaryRedirect},
			"",
			GetHandler,
		},
	}
	config.ParseEnvFlags()

	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := storage.NewStorage()
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
