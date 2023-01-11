package badger

import (
	"fmt"

	"github.com/planetary-social/scuttlego/logging"
)

const originalLogLevelField = "original_level"

type LoggerLevel int

const (
	LoggerLevelError LoggerLevel = iota
	LoggerLevelWarning
	LoggerLevelInfo
	LoggerLevelDebug
)

type Logger struct {
	logger logging.Logger
	level  LoggerLevel
}

func NewLogger(logger logging.Logger, level LoggerLevel) Logger {
	return Logger{logger: logger, level: level}
}

func (l Logger) Errorf(s string, i ...interface{}) {
	if l.level >= LoggerLevelError {
		l.logger.
			WithField(originalLogLevelField, "error").
			Error(fmt.Sprintf(s, i...))
	}
}

func (l Logger) Warningf(s string, i ...interface{}) {
	if l.level >= LoggerLevelWarning {
		l.logger.
			WithField(originalLogLevelField, "warning").
			Error(fmt.Sprintf(s, i...))
	}
}

func (l Logger) Infof(s string, i ...interface{}) {
	if l.level >= LoggerLevelInfo {
		l.logger.
			WithField(originalLogLevelField, "info").
			Debug(fmt.Sprintf(s, i...))
	}
}

func (l Logger) Debugf(s string, i ...interface{}) {
	if l.level >= LoggerLevelDebug {
		l.logger.
			WithField(originalLogLevelField, "debug").
			Debug(fmt.Sprintf(s, i...))
	}
}
