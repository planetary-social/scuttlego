package adapters

import (
	"time"
)

type CurrentTimeProvider struct {
}

func NewCurrentTimeProvider() *CurrentTimeProvider {
	return &CurrentTimeProvider{}
}

func (c CurrentTimeProvider) Get() time.Time {
	return time.Now()
}
