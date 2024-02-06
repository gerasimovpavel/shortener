package cookies

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Cookie(t *testing.T) {
	tests := []struct {
		name        string
		clearcookie bool
		result      bool
	}{
		{"test cookie accept",
			false,
			true},
		{"test cookie accept",
			true,
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookie, _ := NewCookie(nil)
			if tt.clearcookie {
				cookie.Name = ""
			}
			err := cookie.Valid()
			if !assert.Equal(t, tt.result, err == nil) {
				panic(fmt.Errorf("status expect %v actual %v\nerror %v", tt.result, err == nil, err))
			}
		})
	}
}
