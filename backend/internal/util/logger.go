package util

import (
	"strings"

	"github.com/sirupsen/logrus"
)

func NewLogger(environment string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	if strings.EqualFold(environment, "production") {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(logrus.DebugLevel)
	}
	return logger
}
