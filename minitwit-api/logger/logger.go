package logger

import (
	"log"

	"go.uber.org/zap"
)

func InitializeLogger() *zap.SugaredLogger {
	config := zap.NewProductionConfig()
	logger, err := config.Build()
	if err != nil {
		log.Fatal(err)
	}
	return logger.Sugar()
}
