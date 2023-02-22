package tests

import (
	"context"

	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
)

type RawMessageHandlerMock struct {
}

func NewRawMessageHandlerMock() *RawMessageHandlerMock {
	return &RawMessageHandlerMock{}
}

func (r RawMessageHandlerMock) Handle(replicatedFrom identity.Public, msg message.RawMessage) error {
	//TODO implement me
	panic("implement me")
}

type WantedFeedsProviderMock struct {
	GetWantedFeedsReturnValue replication.WantedFeeds
}

func NewWantedFeedsProviderMock() *WantedFeedsProviderMock {
	return &WantedFeedsProviderMock{}
}

func (c WantedFeedsProviderMock) GetWantedFeeds() (replication.WantedFeeds, error) {
	return c.GetWantedFeedsReturnValue, nil
}

type MessageStreamerMock struct {
}

func NewMessageStreamerMock() *MessageStreamerMock {
	return &MessageStreamerMock{}
}

func (m MessageStreamerMock) Handle(ctx context.Context, id refs.Feed, seq *message.Sequence, messageWriter ebt.MessageWriter) {
	//TODO implement me
	panic("implement me")
}
