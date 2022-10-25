package replication

import (
	"sort"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

const (
	// For how long the social graph will be cached before rebuilding it. Until
	// this refresh happens newly discovered feeds are not taken into account.
	refreshContactsEvery = 5 * time.Second
)

type WantedFeedsRepository interface {
	GetWantedFeeds() (WantedFeeds, error)
}

type Contact struct {
	Who       refs.Feed
	Hops      graph.Hops
	FeedState FeedState
}

type WantedFeed struct {
	Who       refs.Feed
	FeedState FeedState
}

// WantedFeeds contains contacts and other wanted feeds. Contacts are sorted by
// hops, ascending. Contacts include the local feed. Other feeds include other
// feeds which should be replicated. Other feeds may include the same feeds.
type WantedFeeds struct {
	contacts   []Contact
	otherFeeds []WantedFeed
}

func NewWantedFeeds(contacts []Contact, otherFeeds []WantedFeed) WantedFeeds {
	sort.Slice(contacts, func(i, j int) bool {
		return contacts[i].Hops.Int() < contacts[j].Hops.Int()
	})

	return WantedFeeds{contacts: contacts, otherFeeds: otherFeeds}
}

func (w WantedFeeds) Contacts() []Contact {
	return w.contacts
}

func (w WantedFeeds) OtherFeeds() []WantedFeed {
	return w.otherFeeds
}

type WantedFeedsCache struct {
	repo WantedFeedsRepository

	cache          []Contact
	cacheTimestamp time.Time
	cacheLock      sync.Mutex // locks cache and cacheTimestamp
}

func NewWantedFeedsCache(repo WantedFeedsRepository) *WantedFeedsCache {
	return &WantedFeedsCache{repo: repo}
}

func (c *WantedFeedsCache) GetContacts() ([]Contact, error) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	if time.Since(c.cacheTimestamp) > refreshContactsEvery {
		v, err := c.repo.GetWantedFeeds()
		if err != nil {
			return nil, errors.Wrap(err, "could not get fresh data")
		}

		// todo this is a hack because I don't want to touch the replicators now, this has to be changed later so that it doesn't create fake contacts
		var contacts []Contact
		for _, feed := range v.OtherFeeds() {
			contacts = append(contacts, Contact{
				Who:       feed.Who,
				Hops:      graph.MustNewHops(1),
				FeedState: feed.FeedState,
			})

		}
		contacts = append(contacts, v.Contacts()...)

		c.cache = contacts
		c.cacheTimestamp = time.Now()
	}

	return c.cache, nil
}
