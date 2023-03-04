package logging

import (
	"context"
)

const (
	loggerFieldName  = "name"
	loggerFieldError = "error"
)

type Level int

const (
	LevelTrace Level = iota
	LevelDebug
	LevelError
	LevelDisabled
)

type Logger interface {
	New(name string) Logger
	WithCtx(ctx context.Context) Logger
	WithError(err error) Logger
	WithField(key string, v any) Logger

	Error() Entry
	Debug() Entry
	Trace() Entry
}

type Entry interface {
	WithError(err error) Entry
	WithField(key string, v any) Entry
	Message(msg string)
}

type LoggingSystem interface {
	EnabledLevel() Level
	Error() LoggingSystemEntry
	Debug() LoggingSystemEntry
	Trace() LoggingSystemEntry
}

type LoggingSystemEntry interface {
	WithField(key string, v any) LoggingSystemEntry
	Message(msg string)
}

type ContextLogger struct {
	ctx    context.Context
	fields map[string]any

	logger LoggingSystem
}

func NewContextLogger(logger LoggingSystem, name string) Logger {
	if logger.EnabledLevel() >= LevelDisabled {
		return NewDevNullLogger()
	}
	return newContextLogger(logger, nil, map[string]any{loggerFieldName: name})
}

func newContextLogger(logger LoggingSystem, ctx context.Context, fields map[string]any) ContextLogger {
	newLogger := ContextLogger{
		ctx:    ctx,
		fields: make(map[string]any),

		logger: logger,
	}

	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

func (l ContextLogger) Error() Entry {
	if l.logger.EnabledLevel() > LevelError {
		return newDevNullLoggerEntry()
	}
	return l.withFields(newEntry(l.logger.Error()))
}

func (l ContextLogger) Debug() Entry {
	if l.logger.EnabledLevel() > LevelDebug {
		return newDevNullLoggerEntry()
	}
	return l.withFields(newEntry(l.logger.Debug()))
}

func (l ContextLogger) Trace() Entry {
	if l.logger.EnabledLevel() > LevelTrace {
		return newDevNullLoggerEntry()
	}
	return l.withFields(newEntry(l.logger.Trace()))
}

func (l ContextLogger) New(name string) Logger {
	newLogger := newContextLogger(l.logger, l.ctx, l.fields)
	v, okExists := l.fields[loggerFieldName]
	if okExists {
		if stringV, okType := v.(string); okType {
			newLogger.fields[loggerFieldName] = stringV + "." + name
			return newLogger
		}
		return newLogger
	}
	newLogger.fields[loggerFieldName] = name
	return newLogger
}

func (l ContextLogger) WithCtx(ctx context.Context) Logger {
	return newContextLogger(l.logger, ctx, l.fields)
}

func (l ContextLogger) WithError(err error) Logger {
	newLogger := newContextLogger(l.logger, l.ctx, l.fields)
	newLogger.fields[loggerFieldError] = err
	return newLogger
}

func (l ContextLogger) WithField(key string, v any) Logger {
	newLogger := newContextLogger(l.logger, l.ctx, l.fields)
	newLogger.fields[key] = v
	return newLogger
}

func (l ContextLogger) withFields(entry Entry) Entry {
	for k, v := range l.fields {
		entry = entry.WithField(k, v)
	}
	entry = addContextFields(entry, l.ctx)
	return entry
}

type entry struct {
	loggingSystemEntry LoggingSystemEntry
}

func newEntry(loggingSystemEntry LoggingSystemEntry) entry {
	return entry{loggingSystemEntry: loggingSystemEntry}
}

func (e entry) WithError(err error) Entry {
	return newEntry(e.loggingSystemEntry.WithField(loggerFieldError, err))
}

func (e entry) WithField(key string, v any) Entry {
	return newEntry(e.loggingSystemEntry.WithField(key, v))
}

func (e entry) Message(msg string) {
	e.loggingSystemEntry.Message(msg)
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

func (d DevNullLogger) Error() Entry {
	return newDevNullLoggerEntry()
}

func (d DevNullLogger) Debug() Entry {
	return newDevNullLoggerEntry()
}

func (d DevNullLogger) Trace() Entry {
	return newDevNullLoggerEntry()
}

type devNullLoggerEntry struct {
}

func newDevNullLoggerEntry() devNullLoggerEntry {
	return devNullLoggerEntry{}
}

func (d devNullLoggerEntry) WithError(err error) Entry {
	return d
}

func (d devNullLoggerEntry) WithField(key string, v any) Entry {
	return d
}

func (d devNullLoggerEntry) Message(msg string) {
}

func addContextFields(entry Entry, ctx context.Context) Entry {
	if ctx == nil {
		return entry
	}

	for label, value := range GetLoggingContext(ctx) {
		entry = entry.WithField(label, value)
	}

	return entry
}
