package pubsub

import (
	"context"

	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
)

type RawMessagePubSub struct {
	pubsub *GoChannelPubSub[message.RawMessage]
}

func NewRawMessagePubSub() *RawMessagePubSub {
	return &RawMessagePubSub{
		pubsub: NewGoChannelPubSub[message.RawMessage](),
	}
}

func (m *RawMessagePubSub) HandleRawMessage(msg message.RawMessage) error {
	m.pubsub.Publish(msg)
	return nil
}

func (m *RawMessagePubSub) SubscribeToNewRawMessages(ctx context.Context) <-chan message.RawMessage {
	return m.pubsub.Subscribe(ctx)
}
