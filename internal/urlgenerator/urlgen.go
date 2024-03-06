// Package urlgen реализует генерацию коротких ссылок
package urlgen

import (
	"github.com/brianvoe/gofakeit"
)

// GenShortOptimized Рандом-генератор коротких ссылок (оптимизированная версия)
func GenShortOptimized() string {
	gofakeit.Seed(0)
	s := gofakeit.Password(true, true, false, false, false, 7)
	return s
}
