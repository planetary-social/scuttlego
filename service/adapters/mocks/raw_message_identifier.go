package mocks

import (
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
)

type RawMessageIdentifierMock struct {
}

func NewRawMessageIdentifierMock() *RawMessageIdentifierMock {
	return &RawMessageIdentifierMock{}
}

func (r RawMessageIdentifierMock) IdentifyRawMessage(raw message.RawMessage) (message.Message, error) {
	return fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed()), nil
}
