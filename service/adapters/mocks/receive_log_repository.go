package mocks

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
)

type ReceiveLogRepositoryMock struct {
}

func NewReceiveLogRepositoryMock() *ReceiveLogRepositoryMock {
	return &ReceiveLogRepositoryMock{}
}

func (r ReceiveLogRepositoryMock) Next(lastSeq uint64) ([]message.Message, error) {
	return nil, errors.New("not implemented")
}
