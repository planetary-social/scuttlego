package logging

import (
	"context"
)

const (
	loggerFieldName  = "name"
	loggerFieldError = "error"
)

type Logger interface {
	New(name string) Logger
	WithCtx(ctx context.Context) Logger
	WithError(err error) Logger
	WithField(key string, v any) Logger

	Error(message string)
	Debug(message string)
	Trace(message string)
}

type LoggingSystem interface {
	WithField(key string, v any) LoggingSystem
	Error(message string)
	Debug(message string)
	Trace(message string)
}

type ContextLogger struct {
	name   string
	logger LoggingSystem
	ctx    context.Context
}

func newContextLogger(logger LoggingSystem, name string, ctx context.Context) Logger {
	return ContextLogger{
		name:   name,
		logger: logger,
		ctx:    ctx,
	}
}

func NewContextLogger(logger LoggingSystem, name string) Logger {
	return newContextLogger(logger, name, nil)
}

func (l ContextLogger) New(name string) Logger {
	return newContextLogger(l.logger, l.name+"."+name, l.ctx)
}

func (l ContextLogger) WithCtx(ctx context.Context) Logger {
	return newContextLogger(l.logger, l.name, ctx)
}

func (l ContextLogger) WithError(err error) Logger {
	return newContextLogger(l.logger.WithField(loggerFieldError, err), l.name, l.ctx)
}

func (l ContextLogger) WithField(key string, v any) Logger {
	return newContextLogger(l.logger.WithField(key, v), l.name, l.ctx)
}

func (l ContextLogger) Error(message string) {
	l.withFields().Error(message)
}

func (l ContextLogger) Debug(message string) {
	l.withFields().Debug(message)
}

func (l ContextLogger) Trace(message string) {
	l.withFields().Trace(message)
}

func (l ContextLogger) withFields() LoggingSystem {
	logger := l.logger.WithField(loggerFieldName, l.name)
	logger = addContextFields(logger, l.ctx)
	return logger
}

type DevNullLogger struct {
}

func NewDevNullLogger() DevNullLogger {
	return DevNullLogger{}
}

func (d DevNullLogger) New(name string) Logger {
	return d
}

func (d DevNullLogger) WithCtx(ctx context.Context) Logger {
	return d
}

func (d DevNullLogger) WithError(err error) Logger {
	return d
}

func (d DevNullLogger) WithField(key string, v any) Logger {
	return d
}

func (d DevNullLogger) Error(message string) {
}

func (d DevNullLogger) Debug(message string) {
}

func (d DevNullLogger) Trace(message string) {
}

func addContextFields(logger LoggingSystem, ctx context.Context) LoggingSystem {
	if ctx == nil {
		return logger
	}

	for label, value := range GetLoggingContext(ctx) {
		logger = logger.WithField(label, value)
	}

	return logger
}
