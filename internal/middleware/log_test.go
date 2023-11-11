package middleware

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var TestLogger *zap.Logger

func Test_Log(t *testing.T) {
	TestLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	Sugar = *TestLogger.Sugar()
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name       string
		wantStatus int
	}{
		{"test logger",
			http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if TestLogger == nil {
				panic(errors.New("logger not created"))
			}
			body := ""
			r := strings.NewReader(body)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", r)

			w := httptest.NewRecorder()
			h := Logger(TestLogger)(http.HandlerFunc(EmptyHandlerFunc))

			h.ServeHTTP(w, req)

			res := w.Result()
			res.Body.Close()

			if !assert.Equal(t, tt.wantStatus, res.StatusCode) {
				panic(fmt.Errorf("status expect %v actual %v\nbody %v", tt.wantStatus, res.StatusCode, body))
			}
		})
	}
}
