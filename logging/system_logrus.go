package logging

import (
	"github.com/sirupsen/logrus"
)

type LogrusLoggingSystem struct {
	logger *logrus.Logger
}

func NewLogrusLoggingSystem(logger *logrus.Logger) LogrusLoggingSystem {
	return LogrusLoggingSystem{
		logger: logger,
	}
}

func (t LogrusLoggingSystem) EnabledLevel() Level {
	switch t.logger.GetLevel() {
	case logrus.TraceLevel:
		return LevelTrace
	case logrus.DebugLevel:
		return LevelDebug
	default:
		return LevelError
	}
}

func (l LogrusLoggingSystem) Error() LoggingSystemEntry {
	return newLogrusEntry(l.logger, logrusLogError)
}

func (l LogrusLoggingSystem) Debug() LoggingSystemEntry {
	return newLogrusEntry(l.logger, logrusLogDebug)
}

func (l LogrusLoggingSystem) Trace() LoggingSystemEntry {
	return newLogrusEntry(l.logger, logrusLogTrace)
}

type logrusEntry struct {
	logger logrus.Ext1FieldLogger
	fields map[string]any
	logFn  logrusLogFn
}

func newLogrusEntry(logger logrus.Ext1FieldLogger, logFn logrusLogFn) *logrusEntry {
	return &logrusEntry{
		logger: logger,
		fields: make(map[string]any),
		logFn:  logFn,
	}
}

func (l logrusEntry) WithField(key string, v any) LoggingSystemEntry {
	entry := newLogrusEntry(l.logger, l.logFn)
	for k, v := range l.fields {
		entry.fields[k] = v
	}
	entry.fields[key] = v
	return entry
}

func (l logrusEntry) Message(msg string) {
	logger := l.logger
	for k, v := range l.fields {
		logger = logger.WithField(k, v)
	}
	l.logFn(logger, msg)
}

type logrusLogFn func(logrus.Ext1FieldLogger, string)

func logrusLogError(logger logrus.Ext1FieldLogger, s string) {
	logger.Error(s)
}

func logrusLogDebug(logger logrus.Ext1FieldLogger, s string) {
	logger.Debug(s)
}

func logrusLogTrace(logger logrus.Ext1FieldLogger, s string) {
	logger.Trace(s)
}
