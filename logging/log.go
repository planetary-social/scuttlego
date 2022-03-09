package logging

import "github.com/sirupsen/logrus"

type Logger interface {
	New(name string) Logger
	WithError(err error) Logger
	WithField(key string, v interface{}) Logger

	Error(message string)
	Debug(message string)
	Trace(message string)
}

type Level int

const (
	LevelError Level = iota
	LevelDebug
	LevelTrace
)

type LogrusLogger struct {
	name   string
	logger logrus.FieldLogger
	level  Level
}

func NewLogrusLogger(logger logrus.FieldLogger, name string, level Level) LogrusLogger {
	return LogrusLogger{
		name:   name,
		logger: logger,
		level:  level,
	}
}

func (l LogrusLogger) New(name string) Logger {
	return NewLogrusLogger(l.logger, l.name+"."+name, l.level)
}

func (l LogrusLogger) WithError(err error) Logger {
	return NewLogrusLogger(l.logger.WithError(err), l.name, l.level)
}

func (l LogrusLogger) WithField(key string, v interface{}) Logger {
	return NewLogrusLogger(l.logger.WithField(key, v), l.name, l.level)
}

func (l LogrusLogger) Error(message string) {
	if l.level >= LevelError {
		l.withName().Error(message)
	}
}

func (l LogrusLogger) Debug(message string) {
	if l.level >= LevelDebug {
		l.withName().Debug(message)
	}
}

func (l LogrusLogger) Trace(message string) {
	if l.level >= LevelTrace {
		l.withName().Debug(message)
	}
}

func (l LogrusLogger) withName() logrus.FieldLogger {
	return l.logger.WithField("name", l.name)
}

type DevNullLogger struct {
}

func NewDevNullLogger() DevNullLogger {
	return DevNullLogger{}
}

func (l DevNullLogger) New(name string) Logger {
	return l
}

func (l DevNullLogger) WithError(err error) Logger {
	return l
}

func (l DevNullLogger) WithField(key string, v interface{}) Logger {
	return l
}

func (l DevNullLogger) Error(message string) {
}

func (l DevNullLogger) Debug(message string) {
}

func (l DevNullLogger) Trace(message string) {
}
