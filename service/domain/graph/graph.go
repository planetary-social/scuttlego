package graph

import (
	"sort"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type Storage interface {
	GetContacts(node refs.Identity) ([]*feeds.Contact, error)
}

type SocialGraph struct {
	graph map[string]Hops
}

func NewSocialGraph(local refs.Identity, hops Hops, storage Storage) (*SocialGraph, error) {
	g := &SocialGraph{
		make(map[string]Hops),
	}
	if err := g.load(hops, local, storage); err != nil {
		return nil, errors.Wrap(err, "failed to load the graph")
	}
	return g, nil
}

func (g *SocialGraph) load(maxHops Hops, local refs.Identity, storage Storage) error {
	localBlocks := internal.NewSet[string]()
	currentHops := MustNewHops(0)
	g.graph[local.String()] = currentHops
	return g.depthFirstSearch(maxHops, currentHops, local, localBlocks, storage)
}

func (g *SocialGraph) depthFirstSearch(maxHops Hops, currentHops Hops, contact refs.Identity, localBlocks internal.Set[string], s Storage) error {
	if currentHops.Int() >= maxHops.Int() {
		return nil
	}

	childContacts, err := s.GetContacts(contact)
	if err != nil {
		return errors.Wrap(err, "could not get contacts")
	}

	if currentHops.Int() == 0 {
		for _, childContact := range childContacts {
			if childContact.Blocking() {
				localBlocks.Put(childContact.Target().String())
			}
		}
	}

	currentHops = MustNewHops(currentHops.Int() + 1)

	for _, childContact := range childContacts {
		if !childContact.Following() || childContact.Blocking() || localBlocks.Contains(childContact.Target().String()) {
			continue
		}

		if existingHops, ok := g.graph[childContact.Target().String()]; ok {
			if existingHops.Int() > currentHops.Int() {
				g.graph[childContact.Target().String()] = currentHops
			}
			continue
		}

		g.graph[childContact.Target().String()] = currentHops

		if err := g.depthFirstSearch(maxHops, currentHops, childContact.Target(), localBlocks, s); err != nil {
			return errors.Wrap(err, "recursion failed")
		}
	}

	return nil
}

func (g *SocialGraph) Contacts() []Contact {
	var result []Contact
	for key, distance := range g.graph {
		result = append(result, Contact{
			Id:   refs.MustNewIdentity(key),
			Hops: distance,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Hops.Int() < result[j].Hops.Int()
	})
	return result
}

func (g *SocialGraph) HasContact(contact refs.Identity) bool {
	_, ok := g.graph[contact.String()]
	return ok
}

type Contact struct {
	Id   refs.Identity
	Hops Hops
}

// Hops specify the number of hops in the graph required to reach a contact. Therefore: 0 is yourself, 1 is a
// person you follow, 2 is a person that is followed by a person that you follow, and so on.
type Hops struct {
	n int
}

func NewHops(n int) (Hops, error) {
	if n < 0 {
		return Hops{}, errors.New("hops must be a non-negative number")
	}
	return Hops{n: n}, nil
}

func MustNewHops(n int) Hops {
	hops, err := NewHops(n)
	if err != nil {
		panic(err)
	}
	return hops
}

func (h Hops) Int() int {
	return h.n
}
