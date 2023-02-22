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

type Contact struct {
	who       refs.Feed
	hops      graph.Hops
	feedState FeedState
}

func NewContact(who refs.Feed, hops graph.Hops, feedState FeedState) (Contact, error) {
	if who.IsZero() {
		return Contact{}, errors.New("zero value of who")
	}
	return Contact{who: who, hops: hops, feedState: feedState}, nil
}

func MustNewContact(who refs.Feed, hops graph.Hops, feedState FeedState) Contact {
	v, err := NewContact(who, hops, feedState)
	if err != nil {
		panic(err)
	}
	return v
}

func (c Contact) Who() refs.Feed {
	return c.who
}

func (c Contact) Hops() graph.Hops {
	return c.hops
}

func (c Contact) FeedState() FeedState {
	return c.feedState
}

func (c Contact) IsZero() bool {
	return c.who.IsZero()
}

type WantedFeed struct {
	who       refs.Feed
	feedState FeedState
}

func NewWantedFeed(who refs.Feed, feedState FeedState) (WantedFeed, error) {
	if who.IsZero() {
		return WantedFeed{}, errors.New("zero value of who")
	}
	return WantedFeed{who: who, feedState: feedState}, nil
}

func MustNewWantedFeed(who refs.Feed, feedState FeedState) WantedFeed {
	v, err := NewWantedFeed(who, feedState)
	if err != nil {
		panic(err)
	}
	return v
}

func (w WantedFeed) Who() refs.Feed {
	return w.who
}

func (w WantedFeed) FeedState() FeedState {
	return w.feedState
}

func (w WantedFeed) IsZero() bool {
	return w.who.IsZero()
}

// WantedFeeds contains contacts and other wanted feeds. Contacts are sorted by
// hops, ascending. Contacts include the local feed. Other feeds include other
// feeds which should be replicated. Other feeds may include the same feeds.
type WantedFeeds struct {
	contacts   []Contact
	otherFeeds []WantedFeed
}

func NewWantedFeeds(contacts []Contact, otherFeeds []WantedFeed) (WantedFeeds, error) {
	for _, contact := range contacts {
		if contact.IsZero() {
			return WantedFeeds{}, errors.New("zero value of contact")
		}
	}

	for _, wantedFeeds := range otherFeeds {
		if wantedFeeds.IsZero() {
			return WantedFeeds{}, errors.New("zero value of wanted feed")
		}
	}

	sort.Slice(contacts, func(i, j int) bool {
		return contacts[i].Hops().Int() < contacts[j].Hops().Int()
	})

	return WantedFeeds{contacts: contacts, otherFeeds: otherFeeds}, nil
}

func MustNewWantedFeeds(contacts []Contact, otherFeeds []WantedFeed) WantedFeeds {
	v, err := NewWantedFeeds(contacts, otherFeeds)
	if err != nil {
		panic(err)
	}
	return v
}

func (w WantedFeeds) Contacts() []Contact {
	return w.contacts
}

func (w WantedFeeds) OtherFeeds() []WantedFeed {
	return w.otherFeeds
}

type WantedFeedsProvider interface {
	GetWantedFeeds() (WantedFeeds, error)
}

type WantedFeedsCache struct {
	provider WantedFeedsProvider

	cache          []Contact
	cacheTimestamp time.Time
	cacheLock      sync.Mutex // locks cache and cacheTimestamp
}

func NewWantedFeedsCache(provider WantedFeedsProvider) *WantedFeedsCache {
	return &WantedFeedsCache{provider: provider}
}

func (c *WantedFeedsCache) GetContacts() ([]Contact, error) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	if time.Since(c.cacheTimestamp) > refreshContactsEvery {
		v, err := c.provider.GetWantedFeeds()
		if err != nil {
			return nil, errors.Wrap(err, "could not get fresh data")
		}

		// todo this is a hack because I don't want to touch the replicators now, this has to be changed later so that it doesn't create fake contacts
		// https://github.com/planetary-social/scuttlego/issues/106
		var contacts []Contact
		for _, wantedFeed := range v.OtherFeeds() {
			contact, err := NewContact(
				wantedFeed.Who(),
				graph.MustNewHops(1),
				wantedFeed.FeedState(),
			)
			if err != nil {
				return nil, errors.Wrap(err, "error creating fake contact")
			}
			contacts = append(contacts, contact)
		}
		contacts = append(contacts, v.Contacts()...)
		// todo contacts can contain dupes?

		c.cache = contacts
		c.cacheTimestamp = time.Now()
	}

	return c.cache, nil
}
