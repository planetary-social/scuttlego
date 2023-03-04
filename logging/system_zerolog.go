package logging

import (
	"github.com/rs/zerolog"
)

type ZerologLoggingSystem struct {
	logger zerolog.Logger
}

func NewZerologLoggingSystem(logger zerolog.Logger) ZerologLoggingSystem {
	logger.GetLevel()
	return ZerologLoggingSystem{
		logger: logger,
	}
}

func (t ZerologLoggingSystem) EnabledLevel() Level {
	switch t.logger.GetLevel() {
	case zerolog.TraceLevel:
		return LevelTrace
	case zerolog.DebugLevel:
		return LevelDebug
	case zerolog.Disabled:
		return LevelDisabled
	default:
		return LevelError
	}
}

func (l ZerologLoggingSystem) Error() LoggingSystemEntry {
	return newZerologEntry(l.logger.Error())
}

func (l ZerologLoggingSystem) Debug() LoggingSystemEntry {
	return newZerologEntry(l.logger.Error())
}

func (l ZerologLoggingSystem) Trace() LoggingSystemEntry {
	return newZerologEntry(l.logger.Error())
}

type zerologEntry struct {
	event *zerolog.Event
}

func newZerologEntry(event *zerolog.Event) *zerologEntry {
	return &zerologEntry{event: event}
}

func (z zerologEntry) WithField(key string, v any) LoggingSystemEntry {
	return newZerologEntry(z.event.Interface(key, v))
}

func (z zerologEntry) Message(msg string) {
	z.event.Msg(msg)
}
