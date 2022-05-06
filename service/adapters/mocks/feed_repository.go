package mocks

import (
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

type FeedRepositoryMockGetMessagesCall struct {
	Id    refs.Feed
	Seq   *message.Sequence
	Limit *int
}

type FeedRepositoryMock struct {
	GetMessagesCalls       []FeedRepositoryMockGetMessagesCall
	GetMessagesReturnValue []message.Message
	GetMessagesReturnErr   error

	CountReturnValue int
}

func NewFeedRepositoryMock() *FeedRepositoryMock {
	return &FeedRepositoryMock{}
}

func (m *FeedRepositoryMock) Count() (int, error) {
	return m.CountReturnValue, nil
}

func (m *FeedRepositoryMock) GetMessages(id refs.Feed, seq *message.Sequence, limit *int) ([]message.Message, error) {
	m.GetMessagesCalls = append(m.GetMessagesCalls, FeedRepositoryMockGetMessagesCall{Id: id, Seq: seq, Limit: limit})
	return m.GetMessagesReturnValue, m.GetMessagesReturnErr
}
