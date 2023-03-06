package mocks

import (
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type BanListRepositoryMock struct {
	mockedBannedFeeds internal.Set[string]
}

func NewBanListRepositoryMock() *BanListRepositoryMock {
	return &BanListRepositoryMock{
		mockedBannedFeeds: internal.NewSet[string](),
	}
}

func (b BanListRepositoryMock) Mock(feed refs.Feed) {
	b.mockedBannedFeeds.Put(feed.String())
}

func (b BanListRepositoryMock) ContainsFeed(feed refs.Feed) (bool, error) {
	return b.mockedBannedFeeds.Contains(feed.String()), nil
}
