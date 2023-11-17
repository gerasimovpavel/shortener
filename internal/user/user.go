package user

import "github.com/brianvoe/gofakeit"

func NewUserId() string {
	return gofakeit.UUID()
}
