package logging

import (
	"context"
)

type Logger interface {
	New(name string) Logger
	WithCtx(ctx context.Context) Logger // todo maybe remove and just use ...Ctx funcs?
	WithError(err error) Logger
	WithField(key string, v any) Logger

	Error(message string)
	ErrorCtx(ctx context.Context, message string)
	Debug(message string)
	DebugCtx(ctx context.Context, message string)
	Trace(message string)
	TraceCtx(ctx context.Context, message string)
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
	return newContextLogger(l.logger.WithField("error", err), l.name, l.ctx)
}

func (l ContextLogger) WithField(key string, v any) Logger {
	return newContextLogger(l.logger.WithField(key, v), l.name, l.ctx)
}

func (l ContextLogger) Error(message string) {
	l.withContextFields(l.withName(l.logger), l.ctx).Error(message)
}

func (l ContextLogger) ErrorCtx(ctx context.Context, message string) {
	l.withContextFields(l.logger, ctx).Error(message)
}

func (l ContextLogger) Debug(message string) {
	l.withContextFields(l.withName(l.logger), l.ctx).Debug(message)
}

func (l ContextLogger) DebugCtx(ctx context.Context, message string) {
	l.withContextFields(l.logger, ctx).Debug(message)
}

func (l ContextLogger) Trace(message string) {
	l.withContextFields(l.withName(l.logger), l.ctx).Trace(message)
}

func (l ContextLogger) TraceCtx(ctx context.Context, message string) {
	l.withContextFields(l.logger, ctx).Trace(message)
}

func (l ContextLogger) withName(logger LoggingSystem) LoggingSystem {
	return logger.WithField("name", l.name)
}

func (l ContextLogger) withContextFields(logger LoggingSystem, ctx context.Context) LoggingSystem {
	if ctx == nil {
		return logger
	}

	for label, value := range GetLoggingContext(ctx) {
		logger = logger.WithField(label, value)
	}

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

func (d DevNullLogger) ErrorCtx(ctx context.Context, message string) {
}

func (d DevNullLogger) DebugCtx(ctx context.Context, message string) {
}

func (d DevNullLogger) TraceCtx(ctx context.Context, message string) {
}
