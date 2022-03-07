package commands

import (
	"github.com/planetary-social/go-ssb/refs"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds"
)

type TransactionProvider interface {
	Transact(func(adapters Adapters) error) error
}

type Adapters struct {
	Feed        FeedRepository
	SocialGraph SocialGraphRepository
}

type FeedRepository interface {
	// UpdateFeed updates the specified feed by calling the provided function on it. Feed is never nil.
	UpdateFeed(ref refs.Feed, f func(feed *feeds.Feed) (*feeds.Feed, error)) error
}

type SocialGraphRepository interface {
	HasContact(contact refs.Identity) (bool, error)
}
