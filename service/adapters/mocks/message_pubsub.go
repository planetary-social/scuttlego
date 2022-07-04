package mocks

import (
	"context"

	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
)

type MessagePubSubMock struct {
	NewMessagesToSend []message.Message
	CallsCount        int
}

func NewMessagePubSubMock() *MessagePubSubMock {
	return &MessagePubSubMock{}
}

// SubscribeToNewMessages closes the channel after the messages are sent as otherwise testing is annoying.
func (m *MessagePubSubMock) SubscribeToNewMessages(ctx context.Context) <-chan message.Message {
	m.CallsCount++

	ch := make(chan message.Message)

	go func() {
		defer close(ch)

		for _, msg := range m.NewMessagesToSend {
			select {
			case ch <- msg:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}
