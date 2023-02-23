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
			childHops := MustNewHops(current.Hops.Int() + 1)

			if childHops.Int() >= maxHops.Int()+1 {
				continue
			}

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

			g.graph[childContact.Target().String()] = childHops
			queue.Enqueue(feedWithDistance{Feed: childContact.Target(), Hops: childHops})
		}
	}

	return g, nil
}

//func (b *SocialGraphBuilder) breadthFirstSearch(
//	g SocialGraph,
//	maxHops Hops,
//	currentHops Hops,
//	contact refs.Identity,
//	localBlocks internal.Set[string],
//) error {
//	if currentHops.Int() >= maxHops.Int() {
//		return nil
//	}
//
//	childContacts, err := b.storage.GetContacts(contact)
//	if err != nil {
//		return errors.Wrap(err, "could not get contacts")
//	}
//
//	if currentHops.Int() == 0 {
//		for _, childContact := range childContacts {
//			if childContact.Blocking() {
//				localBlocks.Put(childContact.Target().String())
//			}
//		}
//	}
//
//	currentHops = MustNewHops(currentHops.Int() + 1)
//
//	for _, childContact := range childContacts {
//		isInBanList, err := b.banList.ContainsFeed(childContact.Target().MainFeed())
//		if err != nil {
//			return errors.Wrap(err, "error checking the ban list")
//		}
//
//		if isInBanList || !childContact.Following() || childContact.Blocking() || localBlocks.Contains(childContact.Target().String()) {
//			continue
//		}
//
//		if existingHops, ok := g.graph[childContact.Target().String()]; ok {
//			if existingHops.Int() > currentHops.Int() {
//				g.graph[childContact.Target().String()] = currentHops
//			}
//			continue
//		}
//
//		g.graph[childContact.Target().String()] = currentHops
//
//		if err := b.depthFirstSearch(g, maxHops, currentHops, childContact.Target(), localBlocks); err != nil {
//			return errors.Wrap(err, "recursion failed")
//		}
//	}
//
//	return nil
//}

type feedWithDistance struct {
	Feed refs.Identity
	Hops Hops
}
