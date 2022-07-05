package mocks

import "time"

type CurrentTimeProviderMock struct {
	CurrentTime time.Time
}

func NewCurrentTimeProviderMock() *CurrentTimeProviderMock {
	return &CurrentTimeProviderMock{}
}

func (c CurrentTimeProviderMock) Get() time.Time {
	if c.CurrentTime.IsZero() {
		return time.Now()
	}
	return c.CurrentTime
}
