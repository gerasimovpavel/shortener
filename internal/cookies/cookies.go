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

	userIDEncrypted, err := crypt.Encrypt(user.NewUserID())
	if err != nil {
		cookie.Name = ""
		return cookie, err
	}
	cookie.Value = userIDEncrypted
	return cookie, nil
}
