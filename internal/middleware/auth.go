// Package middleware реализует посредника для http запросов для авторизации пользователя
package middleware

import (
	"github.com/gerasimovpavel/shortener.git/pkg/cookies"
	"github.com/gerasimovpavel/shortener.git/pkg/crypt"
	"net/http"
)

// UserID ID пользователя
var UserID string

// AuthCookie Проверка авторизации пользователя по куки
func AuthCookie(next http.Handler) http.Handler {
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

// AuthHeader Проверка авторизации пользователя по Header
func AuthHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		UserID = ""

		header := r.Header.Get("Authorization")

		if header == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		UserID, err = crypt.Decrypt(header)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AutoAuthHeader Проверка авторизации пользователя по Header с автоматической авторизацией
func AutoAuthHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		UserID = ""

		header := r.Header.Get("Authorization")

		if header == "" {
			cookie, _ := r.Cookie("UserID")
			err = cookie.Valid()
			if err != nil {
				cookie, err = cookies.NewCookie(cookie)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				http.SetCookie(w, cookie)
			}
			w.Header().Set("Authorization", cookie.Value)
			header = w.Header().Get("Authorization")
		}

		if header == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		UserID, err = crypt.Decrypt(header)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}
