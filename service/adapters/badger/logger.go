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
	logger logging.LoggingSystem
	level  LoggerLevel
}

func NewLogger(logger logging.LoggingSystem, level LoggerLevel) Logger {
	return Logger{logger: logger, level: level}
}

func (l Logger) Errorf(s string, i ...interface{}) {
	if l.level >= LoggerLevelError {
		l.logger.
			Error().
			WithField(originalLogLevelField, "error").
			Message(fmt.Sprintf(s, i...))
	}
}

func (l Logger) Warningf(s string, i ...interface{}) {
	if l.level >= LoggerLevelWarning {
		l.logger.
			Error().
			WithField(originalLogLevelField, "warning").
			Message(fmt.Sprintf(s, i...))
	}
}

func (l Logger) Infof(s string, i ...interface{}) {
	if l.level >= LoggerLevelInfo {
		l.logger.
			Debug().
			WithField(originalLogLevelField, "info").
			Message(fmt.Sprintf(s, i...))
	}
}

func (l Logger) Debugf(s string, i ...interface{}) {
	if l.level >= LoggerLevelDebug {
		l.logger.
			Debug().
			WithField(originalLogLevelField, "debug").
			Message(fmt.Sprintf(s, i...))
	}
}
