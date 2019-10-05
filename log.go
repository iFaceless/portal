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
