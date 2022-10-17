package mocks

import (
	"context"
	"sync/atomic"

	"github.com/planetary-social/scuttlego/service/adapters/pubsub"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
)

type MessagePubSubMock struct {
	subscribeToNewMessagesCallsCount atomic.Int32
	pubsub                           *pubsub.MessagePubSub
}

func NewMessagePubSubMock(pubsub *pubsub.MessagePubSub) *MessagePubSubMock {
	return &MessagePubSubMock{pubsub: pubsub}
}

func (m *MessagePubSubMock) SubscribeToNewMessages(ctx context.Context) <-chan message.Message {
	m.subscribeToNewMessagesCallsCount.Add(1)
	return m.pubsub.SubscribeToNewMessages(ctx)
}

func (m *MessagePubSubMock) SubscribeToNewMessagesCallsCount() int {
	return int(m.subscribeToNewMessagesCallsCount.Load())
}

func (m *MessagePubSubMock) PublishNewMessage(msg message.Message) {
	m.pubsub.PublishNewMessage(msg)
}
