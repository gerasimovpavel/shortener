package middleware

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

// Sugar Логгер от zap
var Sugar zap.SugaredLogger

type (
	responseData struct {
		status int
		size   int
		body   []byte
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write Записб данных
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader Запись заголовка запроса
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

// Logger Запись запросов в лог
func Logger(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			start := time.Now()
			uri := r.RequestURI
			method := r.Method

			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
				responseData:   responseData,
			}

			h.ServeHTTP(&lw, r)

			duration := time.Since(start)

			// отправляем сведения о запросе в zap
			Sugar.Infoln(
				"uri", uri,
				"method", method,
				"duration", duration,
				"status", responseData.status,
				"size", responseData.size,
			)

		})
	}

}
