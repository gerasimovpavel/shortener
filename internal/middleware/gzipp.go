// Package middleware /gzipp реализует посредника для обработки http запросов для сжатия/распаковки данных
package middleware

import (
	"github.com/gerasimovpavel/shortener.git/pkg/compressor"
	"net/http"
	"strings"
)

// Gzip Сжатие данных
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		nw := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {

			cw := compressor.NewCompressWriter(w)

			nw = cw

			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {

			cr, err := compressor.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = cr
			defer cr.Close()
		}

		next.ServeHTTP(nw, r)
	})
}
