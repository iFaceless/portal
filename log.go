//+build !test

package portal

import (
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func SetDebug(v bool) {
	if v {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
}

type LogLevel = logrus.Level

const (
	ErrorLevel = logrus.ErrorLevel
	WarnLevel  = logrus.WarnLevel
	InfoLevel  = logrus.InfoLevel
	DebugLevel = logrus.DebugLevel
)

// SetLogLevel block the logs who's level is lower
func SetLogLevel(level LogLevel) {
	logger.SetLevel(level)
}
