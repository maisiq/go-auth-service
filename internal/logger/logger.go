package logger

import (
	"go.uber.org/zap"
)

var L *zap.SugaredLogger

func InitLogger(debug bool) *zap.SugaredLogger {
	if debug {
		logger, _ := zap.NewDevelopment()
		L = logger.Sugar()
	} else {
		logger, _ := zap.NewProduction()
		L = logger.Sugar()
	}

	return L
}

func GetLogger() *zap.SugaredLogger {
	if L == nil {
		panic("logger is not initialized")
	}
	return L
}
