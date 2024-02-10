package router

import (
	"errors"
	"fmt"
	"github.com/gerasimovpavel/shortener.git/pkg/logger"
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
			err := logger.NewLogger()
			if err != nil {
				panic(fmt.Errorf("failed to create logger: %w", err))
			}
			r := MainRouter()
			if r == nil {
				panic(errors.New("failed to create main router"))
			}
		})
	}
}
