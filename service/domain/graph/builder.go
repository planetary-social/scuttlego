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
	maxHops Hops
	local   refs.Identity

	graph       SocialGraph
	localBlocks internal.Set[string]
	queue       *internal.Queue[feedWithDistance]
}

func NewSocialGraphBuilder(
	storage ContactsStorage,
	banList BanList,
	hops Hops,
	local refs.Identity,
) *SocialGraphBuilder {
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

	return &SocialGraphBuilder{
		storage: storage,
		banList: banList,
		maxHops: hops,
		local:   local,

		graph:       g,
		queue:       queue,
		localBlocks: localBlocks,
	}
}

func (b *SocialGraphBuilder) HasContact(contact refs.Identity) (bool, error) {
	if err := b.buildUntil(&contact); err != nil {
		return false, errors.Wrap(err, "error building the graph")
	}
	return b.graph.HasContact(contact), nil
}

func (b *SocialGraphBuilder) Build() (SocialGraph, error) {
	if err := b.buildUntil(nil); err != nil {
		return SocialGraph{}, errors.Wrap(err, "error building the graph")
	}
	return b.graph, nil
}

func (b *SocialGraphBuilder) buildUntil(buildUntil *refs.Identity) error {
	for {
		if buildUntil != nil {
			if b.graph.HasContact(*buildUntil) {
				break
			}
		}

		current, ok := b.queue.Dequeue()
		if !ok {
			break
		}

		childContacts, err := b.storage.GetContacts(current.Feed)
		if err != nil {
			return errors.Wrap(err, "could not get contacts")
		}

		if current.Hops.Int() == 0 {
			for _, childContact := range childContacts {
				if childContact.Blocking() {
					b.localBlocks.Put(childContact.Target().String())
				}
			}
		}

		for _, childContact := range childContacts {
			if _, visited := b.graph.graph[childContact.Target().String()]; visited {
				continue
			}

			isInBanList, err := b.banList.ContainsFeed(childContact.Target().MainFeed())
			if err != nil {
				return errors.Wrap(err, "error checking the ban list")
			}

			if isInBanList || !childContact.Following() || childContact.Blocking() || b.localBlocks.Contains(childContact.Target().String()) {
				continue
			}

			childHops := MustNewHops(current.Hops.Int() + 1)
			b.graph.graph[childContact.Target().String()] = childHops

			if childHops.Int() < b.maxHops.Int() {
				b.queue.Enqueue(feedWithDistance{Feed: childContact.Target(), Hops: childHops})
			}
		}
	}

	return nil
}

type feedWithDistance struct {
	Feed refs.Identity
	Hops Hops
}
