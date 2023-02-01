package logging

type DevNullLoggingSystem struct {
}

func NewDevNullLoggingSystem() DevNullLoggingSystem {
	return DevNullLoggingSystem{}
}

func (d DevNullLoggingSystem) WithField(key string, v any) LoggingSystem {
	return d
}

func (d DevNullLoggingSystem) Error(message string) {
}

func (d DevNullLoggingSystem) Debug(message string) {
}

func (d DevNullLoggingSystem) Trace(message string) {
}
