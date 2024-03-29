// Package logger реализует создание логгера
package logger

import "go.uber.org/zap"

// Logger Логгер от zap
var Logger *zap.Logger

// NewLogger создание нового логгера
func NewLogger() error {
	var err error
	Logger, err = zap.NewDevelopment()
	defer Logger.Sync()
	if err != nil {
		return err
	}
	return nil
}
