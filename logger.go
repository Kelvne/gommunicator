package gommunicator

import "go.uber.org/zap"

var logger *zap.Logger

func getLogger() *zap.Logger {
	if logger == nil {
		newLogger, err := zap.NewProduction()
		defer logger.Sync()
		if err != nil {
			return nil
		}

		logger = newLogger
	}

	return logger.Named("gommunicator")
}
