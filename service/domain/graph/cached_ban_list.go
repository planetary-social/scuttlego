package graph

import (
	"encoding/hex"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/bans"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type BanListHasher interface {
	HashForFeed(refs.Feed) (bans.Hash, error)
}

type BanListLister interface {
	List() ([]bans.Hash, error)
}

type CachedBanList struct {
	hasher BanListHasher
	hashes internal.Set[string]
}

func NewCachedBanList(hasher BanListHasher, lister BanListLister) (*CachedBanList, error) {
	hashes, err := lister.List()
	if err != nil {
		return nil, errors.Wrap(err, "error listing the ban list")

	}

	hashesSet := internal.NewSet[string]()
	for _, hash := range hashes {
		hashesSet.Put(hex.EncodeToString(hash.Bytes()))
	}

	return &CachedBanList{hasher: hasher, hashes: hashesSet}, nil
}

func (c CachedBanList) ContainsFeed(feed refs.Feed) (bool, error) {
	hashForFeed, err := c.hasher.HashForFeed(feed)
	if err != nil {
		return false, errors.New("error calculating the ban list hash")
	}

	return c.hashes.Contains(hex.EncodeToString(hashForFeed.Bytes())), nil
}
