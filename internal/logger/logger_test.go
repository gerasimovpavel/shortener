package logger

import "testing"

func TestLogger(t *testing.T) {
	err := NewLogger()
	if err != nil {
		panic(err)
	}
}
