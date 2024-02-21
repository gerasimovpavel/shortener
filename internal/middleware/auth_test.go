package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Auth(t *testing.T) {
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
