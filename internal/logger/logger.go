package logger

import "go.uber.org/zap"

var Logger zap.Logger

func NewLogger() error {
	Logger, err := zap.NewDevelopment()
	defer Logger.Sync()
	if err != nil {
		return err
	}
	return nil
}
