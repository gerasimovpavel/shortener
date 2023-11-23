package router

import (
	"errors"
	"testing"
)

func Test_MainRouter(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"create router"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := MainRouter()
			if r == nil {
				panic(errors.New("failed to create main router"))
			}
		})
	}
}
