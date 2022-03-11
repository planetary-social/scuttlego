package mocks

import (
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

type FeedRepositoryMockCall struct {
	Id    refs.Feed
	Seq   *message.Sequence
	Limit *int
}

type FeedRepositoryMock struct {
	Calls       []FeedRepositoryMockCall
	ReturnValue []message.Message
	ReturnErr   error
}

func NewFeedRepositoryMock() *FeedRepositoryMock {
	return &FeedRepositoryMock{}
}

func (f *FeedRepositoryMock) GetMessages(id refs.Feed, seq *message.Sequence, limit *int) ([]message.Message, error) {
	f.Calls = append(f.Calls, FeedRepositoryMockCall{Id: id, Seq: seq, Limit: limit})
	return f.ReturnValue, f.ReturnErr
}
