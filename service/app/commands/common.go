package commands

import (
	"time"

	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type NewPeerHandler interface {
	HandleNewPeer(peer transport.Peer)
}

type PeerManager interface {
	// Connect instructs the peer manager that it should establish communications with the specified node. The peer
	// manager may ignore this request under specific circumstances e.g. it may avoid establishing duplicate
	// connections to a single identity.
	Connect(remote identity.Public, address network.Address) error

	// EstablishNewConnections instructs the peer manager that it is time to establish new connections so that
	// the specific connections quotas are met.
	EstablishNewConnections() error

	// ProcessNewLocalDiscovery informs the peer manager about a new local discovery.
	ProcessNewLocalDiscovery(remote identity.Public, address network.Address) error
}

type TransactionProvider interface {
	Transact(func(adapters Adapters) error) error
}

type Adapters struct {
	Feed        FeedRepository
	SocialGraph SocialGraphRepository
	WantList    WantListRepository
}

type FeedRepository interface {
	// UpdateFeed updates the specified feed by calling the provided function on it. Feed is never nil.
	UpdateFeed(ref refs.Feed, f func(feed *feeds.Feed) (*feeds.Feed, error)) error
}

type SocialGraphRepository interface {
	GetSocialGraph() (*graph.SocialGraph, error)
}

type WantListRepository interface {
	// Add puts the blob in the want list. If the blob can't be retrieved before
	// the specified point of time it will be removed from the want list.
	Add(id refs.Blob, until time.Time) error
}

type CurrentTimeProvider interface {
	Get() time.Time
}
