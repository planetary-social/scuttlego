package badger

import (
	"fmt"

	"github.com/planetary-social/scuttlego/logging"
)

const originalLogLevelField = "original_level"

type Logger struct {
	logger logging.Logger
}

func NewLogger(logger logging.Logger) Logger {
	return Logger{logger: logger}
}

func (l Logger) Errorf(s string, i ...interface{}) {
	l.logger.
		WithField(originalLogLevelField, "error").
		Error(fmt.Sprintf(s, i...))
}

func (l Logger) Warningf(s string, i ...interface{}) {
	l.logger.
		WithField(originalLogLevelField, "warning").
		Error(fmt.Sprintf(s, i...))
}

func (l Logger) Infof(s string, i ...interface{}) {
	l.logger.
		WithField(originalLogLevelField, "info").
		Debug(fmt.Sprintf(s, i...))
}

func (l Logger) Debugf(s string, i ...interface{}) {
	l.logger.
		WithField(originalLogLevelField, "debug").
		Debug(fmt.Sprintf(s, i...))
}
