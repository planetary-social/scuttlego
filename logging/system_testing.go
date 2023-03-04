package logging

import (
	"fmt"
	"testing"
)

type TestingLoggingSystem struct {
	tb testing.TB
}

func NewTestingLoggingSystem(tb testing.TB) *TestingLoggingSystem {
	return &TestingLoggingSystem{tb: tb}
}

func (t TestingLoggingSystem) EnabledLevel() Level {
	return LevelTrace
}

func (t TestingLoggingSystem) Error() LoggingSystemEntry {
	return newTestingEntry(t.tb.Log)
}

func (t TestingLoggingSystem) Debug() LoggingSystemEntry {
	return newTestingEntry(t.tb.Log)
}

func (t TestingLoggingSystem) Trace() LoggingSystemEntry {
	return newTestingEntry(t.tb.Log)
}

type testingEntry struct {
	log func(args ...any)
}

func newTestingEntry(log func(args ...any)) *testingEntry {
	return &testingEntry{log: log}
}

func (t testingEntry) WithField(key string, v any) LoggingSystemEntry {
	prev := t.log
	return newTestingEntry(func(args ...any) {
		tmp := []any{fmt.Sprintf("%s=%s", key, v)}
		tmp = append(tmp, args...)
		prev(tmp)
	})
}

func (t testingEntry) Message(msg string) {
	t.log(msg)
}
