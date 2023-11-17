package cookies

import (
	"github.com/gerasimovpavel/shortener.git/internal/crypt"
	"github.com/gerasimovpavel/shortener.git/internal/user"
	"net/http"
	"time"
)

func NewCookie(cookie *http.Cookie) (*http.Cookie, error) {
	if cookie == nil {
		cookie = &http.Cookie{}
	}
	cookie.Name = "UserID"
	cookie.Expires = time.Now().Add(time.Hour * 24)
	cookie.Path = "/"

	userIdEncrypted, err := crypt.Encrypt(user.NewUserId())
	if err != nil {
		cookie.Name = ""
		return cookie, err
	}
	cookie.Value = userIdEncrypted
	return cookie, nil
}
