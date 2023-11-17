package user

import "github.com/brianvoe/gofakeit"

func NewUserID() string {
	return gofakeit.UUID()
}
