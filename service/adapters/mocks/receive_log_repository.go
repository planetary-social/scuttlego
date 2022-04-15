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

func (r ReceiveLogRepositoryMock) Get(startSeq int, limit int) ([]message.Message, error) {
	return nil, errors.New("not implemented")
}
