package logging

import (
	"github.com/sirupsen/logrus"
)

type LogrusLoggingSystem struct {
	logger logrus.Ext1FieldLogger
}

func NewLogrusLoggingSystem(logger logrus.Ext1FieldLogger) LogrusLoggingSystem {
	return LogrusLoggingSystem{
		logger: logger,
	}
}

func (l LogrusLoggingSystem) WithField(key string, v any) LoggingSystem {
	return NewLogrusLoggingSystem(l.logger.WithField(key, v))
}

func (l LogrusLoggingSystem) Error(message string) {
	l.logger.Error(message)
}

func (l LogrusLoggingSystem) Debug(message string) {
	l.logger.Debug(message)
}

func (l LogrusLoggingSystem) Trace(message string) {
	l.logger.Trace(message)
}
