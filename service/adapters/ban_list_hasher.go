package adapters

import (
	"crypto/sha256"

	"github.com/planetary-social/scuttlego/service/domain/bans"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type BanListHasher struct {
}

func NewBanListHasher() *BanListHasher {
	return &BanListHasher{}
}

func (b BanListHasher) HashForFeed(feed refs.Feed) (bans.Hash, error) {
	hash := sha256.New()
	hash.Write([]byte(feed.String()))
	sum := hash.Sum(nil)
	return bans.NewHash(sum)
}
