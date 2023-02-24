package graph

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type ContactsStorage interface {
	GetContacts(node refs.Identity) ([]*feeds.Contact, error)
}

type BanList interface {
	ContainsFeed(feed refs.Feed) (bool, error)
}

type SocialGraphBuilder struct {
	storage ContactsStorage
	banList BanList
}

func NewSocialGraphBuilder(storage ContactsStorage, banList BanList) *SocialGraphBuilder {
	return &SocialGraphBuilder{storage: storage, banList: banList}
}

func (b *SocialGraphBuilder) Build(maxHops Hops, local refs.Identity) (SocialGraph, error) {
	g := SocialGraph{
		graph: make(map[string]Hops),
	}
	localBlocks := internal.NewSet[string]()
	queue := internal.NewQueue[feedWithDistance]()

	g.graph[local.String()] = MustNewHops(0)
	queue.Enqueue(feedWithDistance{
		Feed: local,
		Hops: MustNewHops(0),
	})

	for {
		current, ok := queue.Dequeue()
		if !ok {
			break
		}

		childContacts, err := b.storage.GetContacts(current.Feed)
		if err != nil {
			return SocialGraph{}, errors.Wrap(err, "could not get contacts")
		}

		if current.Hops.Int() == 0 {
			for _, childContact := range childContacts {
				if childContact.Blocking() {
					localBlocks.Put(childContact.Target().String())
				}
			}
		}

		for _, childContact := range childContacts {
			if _, visited := g.graph[childContact.Target().String()]; visited {
				continue
			}

			isInBanList, err := b.banList.ContainsFeed(childContact.Target().MainFeed())
			if err != nil {
				return SocialGraph{}, errors.Wrap(err, "error checking the ban list")
			}

			if isInBanList || !childContact.Following() || childContact.Blocking() || localBlocks.Contains(childContact.Target().String()) {
				continue
			}

			childHops := MustNewHops(current.Hops.Int() + 1)
			g.graph[childContact.Target().String()] = childHops

			if childHops.Int() < maxHops.Int() {
				queue.Enqueue(feedWithDistance{Feed: childContact.Target(), Hops: childHops})
			}
		}
	}

	return g, nil
}

type feedWithDistance struct {
	Feed refs.Identity
	Hops Hops
}
