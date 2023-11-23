package logger

import "go.uber.org/zap"

var Logger *zap.Logger

func NewLogger() error {
	var err error
	Logger, err = zap.NewDevelopment()
	defer Logger.Sync()
	if err != nil {
		return err
	}
	return nil
}
