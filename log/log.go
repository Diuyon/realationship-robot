package log

import (
	"github.com/sirupsen/logrus"
)

var logger Logger

func init() {
	logger = logrus.New()
}

type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Debugf(format string, v ...interface{})
	Debugln(v ...interface{})
	Errorf(format string, v ...interface{})
	Errorln(v ...interface{})
}

func Setup(l Logger) {
	logger = l
}

func Printf(format string, v ...interface{}) {
	logger.Printf(format, v...)
}

func Println(v ...interface{}) {
	logger.Println(v...)
}

func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

func Debugln(v ...interface{}) {
	logger.Debugln(v...)
}

func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

func Errorln(v ...interface{}) {
	logger.Errorln(v...)
}
