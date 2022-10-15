package mocks

import (
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type FeedRepositoryMockGetMessagesCall struct {
	Id    refs.Feed
	Seq   *message.Sequence
	Limit *int
}

type FeedRepositoryMockGetMessageCall struct {
	Feed refs.Feed
	Seq  message.Sequence
}

type FeedRepositoryMock struct {
	GetMessagesCalls       []FeedRepositoryMockGetMessagesCall
	GetMessagesReturnValue []message.Message
	GetMessagesReturnErr   error

	GetMessageCalls       []FeedRepositoryMockGetMessageCall
	GetMessageReturnValue message.Message

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

func (m *FeedRepositoryMock) GetMessage(feed refs.Feed, sequence message.Sequence) (message.Message, error) {
	m.GetMessageCalls = append(m.GetMessageCalls, FeedRepositoryMockGetMessageCall{Feed: feed, Seq: sequence})
	return m.GetMessageReturnValue, nil
}
