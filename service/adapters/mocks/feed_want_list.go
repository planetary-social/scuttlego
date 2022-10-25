package mocks

import (
	"time"

	"github.com/boreq/errors"
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

func (f *FeedWantListRepositoryMock) List() ([]refs.Feed, error) {
	return nil, errors.New("not implemented")
}

func (f *FeedWantListRepositoryMock) Contains(id refs.Feed) (bool, error) {
	return false, errors.New("not implemented")
}

type FeedWantListRepositoryMockAddCall struct {
	Id    refs.Feed
	Until time.Time
}
