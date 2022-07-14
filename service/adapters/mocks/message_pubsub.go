package mocks

import (
	"context"

	"github.com/planetary-social/scuttlego/service/adapters/pubsub"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
)

type MessagePubSubMock struct {
	CallsCount int
	pubsub     *pubsub.MessagePubSub
}

func NewMessagePubSubMock(pubsub *pubsub.MessagePubSub) *MessagePubSubMock {
	return &MessagePubSubMock{pubsub: pubsub}
}

func (m *MessagePubSubMock) SubscribeToNewMessages(ctx context.Context) <-chan message.Message {
	m.CallsCount++
	return m.pubsub.SubscribeToNewMessages(ctx)
}

func (m *MessagePubSubMock) PublishNewMessage(msg message.Message) {
	m.pubsub.PublishNewMessage(msg)
}
