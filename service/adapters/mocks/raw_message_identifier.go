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

func (r *RawMessageIdentifierMock) VerifyRawMessage(raw message.RawMessage) (message.Message, error) {
	if msg, ok := r.v[hex.EncodeToString(raw.Bytes())]; ok {
		return msg, nil
	}

	return fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed()), nil
}

func (r *RawMessageIdentifierMock) LoadRawMessage(raw message.VerifiedRawMessage) (message.MessageWithoutId, error) {
	if msg, ok := r.v[hex.EncodeToString(raw.Bytes())]; ok {
		return r.convert(msg)
	}

	return r.convert(fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed()))
}

func (r *RawMessageIdentifierMock) Mock(msg message.Message) {
	key := hex.EncodeToString(msg.Raw().Bytes())
	if _, ok := r.v[key]; ok {
		panic("message with identical raw bytes was already mocked")
	}
	r.v[key] = msg
}

func (r *RawMessageIdentifierMock) convert(msg message.Message) (message.MessageWithoutId, error) {
	return message.NewMessageWithoutId(
		msg.Previous(),
		msg.Sequence(),
		msg.Author(),
		msg.Feed(),
		msg.Timestamp(),
		msg.Content(),
		msg.Raw(),
	)
}
