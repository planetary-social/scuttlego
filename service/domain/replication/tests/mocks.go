package tests

import (
	"context"

	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
)

type RawMessageHandlerMock struct {
}

func NewRawMessageHandlerMock() *RawMessageHandlerMock {
	return &RawMessageHandlerMock{}
}

func (r RawMessageHandlerMock) Handle(msg message.RawMessage) error {
	//TODO implement me
	panic("implement me")
}

type ContactsRepositoryMock struct {
	GetContactsReturnValue []replication.Contact
}

func NewContactsRepositoryMock() *ContactsRepositoryMock {
	return &ContactsRepositoryMock{}
}

func (c ContactsRepositoryMock) GetContacts() ([]replication.Contact, error) {
	return c.GetContactsReturnValue, nil
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
