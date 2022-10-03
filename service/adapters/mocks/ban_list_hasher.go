package mocks

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/bans"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type BanListHasherMock struct {
	hashes map[string]bans.Hash
}

func NewBanListHasherMock() *BanListHasherMock {
	return &BanListHasherMock{
		hashes: make(map[string]bans.Hash),
	}
}

func (b BanListHasherMock) HashForFeed(feed refs.Feed) (bans.Hash, error) {
	h, ok := b.hashes[feed.String()]
	if !ok {
		return bans.Hash{}, errors.New("hash not mocked")
	}
	return h, nil
}

func (b BanListHasherMock) Mock(feed refs.Feed, hash bans.Hash) {
	b.hashes[feed.String()] = hash
}
