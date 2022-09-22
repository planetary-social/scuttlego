package replication

import (
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

type ContactsRepository interface {
	// GetContacts returns a list of contacts. Contacts are sorted by hops,
	// ascending. Contacts include the local feed.
	GetContacts() ([]Contact, error)
}

type Contact struct {
	Who       refs.Feed
	Hops      graph.Hops
	FeedState FeedState
}

type ContactsCache struct {
	repo ContactsRepository

	contactsCache          []Contact
	contactsCacheTimestamp time.Time
	contactsCacheLock      sync.Mutex // locks contactsCache and contactsCacheTimestamp
}

func NewContactsCache(repo ContactsRepository) *ContactsCache {
	return &ContactsCache{repo: repo}
}

func (c *ContactsCache) GetContacts() ([]Contact, error) {
	c.contactsCacheLock.Lock()
	defer c.contactsCacheLock.Unlock()

	if time.Since(c.contactsCacheTimestamp) > refreshContactsEvery {
		contacts, err := c.repo.GetContacts()
		if err != nil {
			return nil, errors.Wrap(err, "could not get fresh contacts")
		}

		c.contactsCache = contacts
		c.contactsCacheTimestamp = time.Now()
	}

	return c.contactsCache, nil
}
