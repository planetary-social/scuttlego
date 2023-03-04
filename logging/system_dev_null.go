package logging

type DevNullLoggingSystem struct {
}

func NewDevNullLoggingSystem() DevNullLoggingSystem {
	return DevNullLoggingSystem{}
}

func (d DevNullLoggingSystem) EnabledLevel() Level {
	return LevelDisabled
}

func (d DevNullLoggingSystem) Error() LoggingSystemEntry {
	return newDevNullEntry()
}

func (d DevNullLoggingSystem) Debug() LoggingSystemEntry {
	return newDevNullEntry()
}

func (d DevNullLoggingSystem) Trace() LoggingSystemEntry {
	return newDevNullEntry()
}

type devNullEntry struct {
}

func newDevNullEntry() devNullEntry {
	return devNullEntry{}
}

func (d devNullEntry) WithField(string, any) LoggingSystemEntry {
	return d
}

func (d devNullEntry) Message(string) {
}
