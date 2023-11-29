package middleware

import (
	"github.com/gerasimovpavel/shortener.git/internal/cookies"
	"github.com/gerasimovpavel/shortener.git/internal/crypt"
	"net/http"
)

var UserID string

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UserID = ""
		cookie, _ := r.Cookie("UserID")
		err := cookie.Valid()
		if err != nil {
			cookie, err = cookies.NewCookie(cookie)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		if cookie.Value == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		UserID, err = crypt.Decrypt(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if cookie.Value == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, cookie)
		next.ServeHTTP(w, r)
	})
}
