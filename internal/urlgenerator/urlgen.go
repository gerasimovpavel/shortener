package urlgen

import (
	"math/rand"
	"time"
)

// GenShort Рандом-генератор коротких ссылок
func GenShort() string {
	const allowchars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const urllen = 7
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	short := make([]byte, urllen)
	for i := range short {
		short[i] = allowchars[seed.Intn(len(allowchars))]
	}
	return string(short)
}
