package logger

import (
	"github.com/sirupsen/logrus"
)

var log = &logrus.Logger{
	Out:   logrus.StandardLogger().Out,
	Level: logrus.DebugLevel,
	Formatter: &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	},
	ReportCaller: true,
}

type Logger struct {
	Module string
	log    *logrus.Logger
}

func NewLogger(module string) *Logger {
	return &Logger{
		Module: module,
		log:    log,
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log.Debugf(format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log.Infof(format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log.Warnf(format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log.Errorf(format, args...)
}
