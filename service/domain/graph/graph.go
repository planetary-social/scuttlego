package graph

import (
	"sort"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

type Storage interface {
	GetContacts(node refs.Identity) ([]refs.Identity, error)
}

type SocialGraph struct {
	graph map[string]Hops // ref.String() -> hops
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

func (g *SocialGraph) load(hops Hops, local refs.Identity, storage Storage) error {
	return g.dfs(hops, 0, local, storage)
}

func (g *SocialGraph) dfs(hops Hops, depth int, node refs.Identity, s Storage) error {
	if depth > hops.Int() {
		return nil
	}

	g.graph[node.String()] = MustNewHops(depth)

	contacts, err := s.GetContacts(node)
	if err != nil {
		return errors.Wrap(err, "could not get contacts")
	}

	for _, contact := range contacts {
		if _, ok := g.graph[contact.String()]; ok {
			continue
		}

		if err := g.dfs(hops, depth+1, contact, s); err != nil {
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
