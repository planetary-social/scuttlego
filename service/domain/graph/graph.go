package graph

import (
	"sort"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type SocialGraph struct {
	graph map[string]Hops
}

func (g SocialGraph) Contacts() []Contact {
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

func (g SocialGraph) HasContact(contact refs.Identity) bool {
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
