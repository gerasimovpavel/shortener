package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_AuthCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	w := httptest.NewRecorder()
	h := AuthCookie(http.HandlerFunc(EmptyHandlerFunc))

	h.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()

	var cookie *http.Cookie
	for _, c := range res.Cookies() {
		if c.Name == "UserID" {
			cookie = c
			break
		}
	}
	err := cookie.Valid()

	if err != nil {
		panic(fmt.Errorf("cookie error: %w", err))
	}

}

func Test_AuthHeader(t *testing.T) {

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string("http://ya.ru")))

	w := httptest.NewRecorder()
	h := AuthHeader(http.HandlerFunc(EmptyHandlerFunc))

	h.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()
	header := res.Header.Get("Authorization")

	if header != "" {
		panic(errors.New("unexpected auth header"))
	}

}

func Test_AuthAutoHeader(t *testing.T) {

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string("http://ya.ru")))

	w := httptest.NewRecorder()
	h := AutoAuthHeader(http.HandlerFunc(EmptyHandlerFunc))

	h.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()
	header := res.Header.Get("Authorization")

	if header == "" {
		panic(errors.New("unexpected empty auth header"))
	}

}
