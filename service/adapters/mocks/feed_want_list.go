package mocks

import (
	"time"

	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type FeedWantListRepositoryMock struct {
	AddCalls []FeedWantListRepositoryMockAddCall
}

func NewFeedWantListRepositoryMock() *FeedWantListRepositoryMock {
	return &FeedWantListRepositoryMock{}
}

func (f *FeedWantListRepositoryMock) Add(id refs.Feed, until time.Time) error {
	f.AddCalls = append(f.AddCalls, FeedWantListRepositoryMockAddCall{
		Id:    id,
		Until: until,
	})
	return nil
}

type FeedWantListRepositoryMockAddCall struct {
	Id    refs.Feed
	Until time.Time
}
