package mocks

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
)

type MarshalerMock struct {
}

func NewMarshalerMock() *MarshalerMock {
	return &MarshalerMock{}
}

func (m MarshalerMock) Marshal(content content.KnownMessageContent) (message.RawMessageContent, error) {
	return message.RawMessageContent{}, errors.New("not implemented")
}

func (m MarshalerMock) Unmarshal(b message.RawMessageContent) (content.KnownMessageContent, error) {
	return content.NewUnknown(b.Bytes())
}
