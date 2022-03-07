package graph

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/identity"
	"github.com/planetary-social/go-ssb/refs"
)

type Storage interface {
	GetContacts(node refs.Identity) ([]refs.Identity, error)
}

type SocialGraph struct {
	graph map[string]int // identity ref -> hops
}

func NewSocialGraph(local identity.Public, storage Storage) (*SocialGraph, error) {
	g := &SocialGraph{
		make(map[string]int),
	}
	if err := g.load(local, storage); err != nil {
		return nil, errors.Wrap(err, "failed to load the graph")
	}
	return g, nil
}

func (g *SocialGraph) load(local identity.Public, storage Storage) error {
	localRef, err := refs.NewIdentityFromPublic(local)
	if err != nil {
		return errors.Wrap(err, "error creating a local ref")
	}

	return g.dfs(0, localRef, storage)
}

func (g *SocialGraph) dfs(depth int, node refs.Identity, s Storage) error {
	g.graph[node.String()] = depth

	contacts, err := s.GetContacts(node)
	if err != nil {
		return errors.Wrap(err, "could not get contacts")
	}

	for _, contact := range contacts {
		if _, ok := g.graph[contact.String()]; ok {
			continue
		}

		if err := g.dfs(depth, contact, s); err != nil {
			return errors.Wrap(err, "recursion failed")
		}
	}

	return nil
}

func (g *SocialGraph) Contacts() []refs.Identity {
	var result []refs.Identity
	for key, distance := range g.graph {
		if g.closeEnough(distance) {
			result = append(result, refs.MustNewIdentity(key))
		}
	}
	return result
}

func (g *SocialGraph) HasContact(contact refs.Identity) bool {
	distance, ok := g.graph[contact.String()]
	// todo solve this differently
	if !g.closeEnough(distance) {
		return false
	}
	return ok
}

func (g *SocialGraph) closeEnough(distance int) bool {
	return distance < 3
}
