// Package cookies реализует управление cookie в http запросах
package cookies

import (
	"github.com/brianvoe/gofakeit"
	"github.com/gerasimovpavel/shortener.git/pkg/crypt"
	"net/http"
	"time"
)

// NewCookie Создание куки с данными аутентификации
func NewCookie(cookie *http.Cookie) (*http.Cookie, error) {
	gofakeit.Seed(0)
	if cookie == nil {
		cookie = &http.Cookie{}
	}
	cookie.Name = "UserID"
	cookie.Expires = time.Now().Add(time.Hour * 24)
	cookie.Path = "/"

	userIDEncrypted, err := crypt.Encrypt(gofakeit.UUID())
	if err != nil {
		cookie.Name = ""
		return cookie, err
	}
	cookie.Value = userIDEncrypted
	return cookie, nil
}
