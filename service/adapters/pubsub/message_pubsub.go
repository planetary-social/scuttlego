package pubsub

import (
	"context"

	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
)

type MessagePubSub struct {
	pubsub *GoChannelPubSub[message.Message]
}

func NewMessagePubSub() *MessagePubSub {
	return &MessagePubSub{
		pubsub: NewGoChannelPubSub[message.Message](),
	}
}

func (m *MessagePubSub) PublishNewMessage(msg message.Message) {
	m.pubsub.Publish(msg)
}

func (m *MessagePubSub) SubscribeToNewMessages(ctx context.Context) <-chan message.Message {
	return m.pubsub.Subscribe(ctx)
}
