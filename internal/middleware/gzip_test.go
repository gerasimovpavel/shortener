package middleware

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func EmptyHandlerFunc(w http.ResponseWriter, r *http.Request) {
}

func Test_Gzip(t *testing.T) {
	tests := []struct {
		name           string
		acceptEncoding string
		contenEncoding string
		compressed     bool
		wantStatus     int
	}{
		{"accept-encoding",
			"gzip",
			"",
			false,
			http.StatusOK},
		{"accept-encoding content-encoding",
			"gzip",
			"gzip",
			true,
			http.StatusOK},
		{"accept-encoding content-encoding error",
			"gzip",
			"gzip",
			false,
			http.StatusInternalServerError},
		{"content-encoding compressed",
			"",
			"gzip",
			true,
			http.StatusOK},
		{"content-encoding error",
			"",
			"gzip",
			false,
			http.StatusInternalServerError},
	}
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			jsonBatch, err := json.Marshal(urls)
			if err != nil {
				panic(err)
			}
			bb := jsonBatch
			if tt.compressed {
				var b bytes.Buffer
				w := gzip.NewWriter(&b)
				w.Write(jsonBatch)
				bb = b.Bytes()
				w.Close()
			}

			body := string(bb)

			r := strings.NewReader(body)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", r)
			req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			req.Header.Set("Content-Encoding", tt.contenEncoding)

			w := httptest.NewRecorder()
			h := Gzip(http.HandlerFunc(EmptyHandlerFunc))

			h.ServeHTTP(w, req)

			res := w.Result()
			res.Body.Close()

			if !assert.Equal(t, tt.wantStatus, res.StatusCode) {
				panic(fmt.Errorf("status expect %v actual %v\nbody %v", tt.wantStatus, res.StatusCode, body))
			}
		})
	}
}
