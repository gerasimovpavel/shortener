package router

import "testing"

func Test_MainRouter(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"create router"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MainRouter()
			if err != nil {
				panic(err)
			}
		})
	}
}
