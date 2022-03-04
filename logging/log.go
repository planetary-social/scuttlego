package logging

import "github.com/sirupsen/logrus"

type Fields map[string]interface{}

type Logger interface {
	New(name string) Logger
	WithError(err error) Logger
	WithField(key string, v interface{}) Logger
	WithFields(fields Fields) Logger

	Error(message string)
	Debug(message string)
}

type LogrusLogger struct {
	name   string
	logger logrus.FieldLogger
}

func NewLogrusLogger(logger logrus.FieldLogger, name string) LogrusLogger {
	return LogrusLogger{
		name:   name,
		logger: logger,
	}
}

func (l LogrusLogger) New(name string) Logger {
	return NewLogrusLogger(l.logger, l.name+"."+name)
}

func (l LogrusLogger) WithError(err error) Logger {
	return NewLogrusLogger(l.logger.WithError(err), l.name)
}

func (l LogrusLogger) WithField(key string, v interface{}) Logger {
	return NewLogrusLogger(l.logger.WithField(key, v), l.name)
}

func (l LogrusLogger) WithFields(fields Fields) Logger {
	return NewLogrusLogger(l.logger.WithFields(logrus.Fields(fields)), l.name)
}

func (l LogrusLogger) Error(message string) {
	l.withName().Error(message)
}

func (l LogrusLogger) Debug(message string) {
	l.withName().Debug(message)
}

func (l LogrusLogger) withName() logrus.FieldLogger {
	return l.logger.WithField("name", l.name)
}
