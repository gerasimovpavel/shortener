package crypt

import (
	"fmt"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_Crypt(t *testing.T) {
	tests := []struct {
		name   string
		msg    string
		result string
	}{
		{"test encrypt",
			"Hello, world",
			"32c5327350d9faaf6ce0dc50bd40ba2f85c6e2f8e846dfc804fefa85"},
		{"test decrypt",
			"32c5327350d9faaf6ce0dc50bd40ba2f85c6e2f8e846dfc804fefa85",
			"Hello, world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res string
			var err error

			config.Cfg.PassphraseKey = "LH;bjdsahlbhfu"

			if strings.Contains(tt.name, "encrypt") {
				res, err = Encrypt(tt.msg)
			}
			if strings.Contains(tt.name, "decrypt") {
				res, err = Decrypt(tt.msg)
			}

			if !assert.Equal(t, tt.result, res) {
				panic(fmt.Errorf("status expect %v actual %v\n %w", tt.result, res, err))
			}
		})
	}
}
