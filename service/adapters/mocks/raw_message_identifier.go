package mocks

import (
	"encoding/hex"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
)

type RawMessageIdentifierMock struct {
	v map[string]message.Message
}

func NewRawMessageIdentifierMock() *RawMessageIdentifierMock {
	return &RawMessageIdentifierMock{
		v: make(map[string]message.Message),
	}
}

func (r *RawMessageIdentifierMock) IdentifyRawMessage(raw message.RawMessage) (message.Message, error) {
	if msg, ok := r.v[hex.EncodeToString(raw.Bytes())]; ok {
		return msg, nil
	}

	return fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed()), nil
}

func (r *RawMessageIdentifierMock) Mock(msg message.Message) {
	r.v[hex.EncodeToString(msg.Raw().Bytes())] = msg
}
