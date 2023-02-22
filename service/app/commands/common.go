package commands

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/bans"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type PeerManager interface {
	// Connect instructs the peer manager that it should establish
	// communications with the specified node. The peer manager may ignore this
	// request under specific circumstances e.g. it may avoid establishing
	// duplicate connections to a single identity.
	Connect(ctx context.Context, remote identity.Public, address network.Address) error

	// ConnectViaRoom instructs the peer manager that it should establish
	// communications with the specified node using a room as a relay. Behaves
	// like Connect.
	ConnectViaRoom(ctx context.Context, portal transport.Peer, target identity.Public) error

	// EstablishNewConnections instructs the peer manager that it is time to
	// establish new connections so that the specific connections quotas are
	// met.
	EstablishNewConnections(ctx context.Context) error

	// ProcessNewLocalDiscovery informs the peer manager about a new local
	// discovery.
	ProcessNewLocalDiscovery(ctx context.Context, remote identity.Public, address network.Address) error

	// DisconnectAll disconnects all peers.
	DisconnectAll() error

	TrackPeer(ctx context.Context, peer transport.Peer)
}

type TransactionProvider interface {
	Transact(func(adapters Adapters) error) error
}

type ReceiveLogRepository interface {
	PutUnderSpecificSequence(id refs.Message, sequence common.ReceiveLogSequence) error

	// ReserveSequencesUpTo ensures that sequences all the way to and including
	// the provided sequence will not be used for automatic sequence generation.
	ReserveSequencesUpTo(sequence common.ReceiveLogSequence) error

	// GetMessage returns the message that the provided receive log sequence
	// points to. Returns common.ErrReceiveLogEntryNotFound if not found.
	GetMessage(seq common.ReceiveLogSequence) (message.Message, error)
}

type Adapters struct {
	Feed         FeedRepository
	ReceiveLog   ReceiveLogRepository
	SocialGraph  SocialGraphRepository
	BlobWantList BlobWantListRepository
	FeedWantList FeedWantListRepository
	BanList      BanListRepository
}

type UpdateFeedFn func(feed *feeds.Feed) error

type FeedRepository interface {
	// UpdateFeed updates the specified feed by calling the provided function on
	// it. Feed is never nil.
	UpdateFeed(ref refs.Feed, f UpdateFeedFn) error

	// UpdateFeedIgnoringReceiveLog works like UpdateFeed but doesn't put
	// messages in receive log.
	UpdateFeedIgnoringReceiveLog(ref refs.Feed, f UpdateFeedFn) error

	// DeleteFeed removes the feed with all associated data.
	DeleteFeed(ref refs.Feed) error

	// GetMessage returns a message with a given sequence from the specified
	// feed.
	GetMessage(ref refs.Feed, sequence message.Sequence) (message.Message, error)

	// RemoveMessagesAtOrAboveSequence removes all feed messages with sequence
	// greater or equal to the given one.
	RemoveMessagesAtOrAboveSequence(ref refs.Feed, sequence message.Sequence) error
}

type SocialGraphRepository interface {
	GetSocialGraph() (graph.SocialGraph, error)
}

type BlobWantListRepository interface {
	// Add puts the blob in the want list. If the blob can't be retrieved before
	// the specified point of time it will be removed from the want list.
	Add(id refs.Blob, until time.Time) error
}

// FeedWantListRepository adds a way to temporarily add feeds to the list of
// replicated feeds. Those feeds will be replicated even if they are not in the
// social graph.
type FeedWantListRepository interface {
	// Add puts the feed in the want list. The entry is removed after the
	// specified amount of time.
	Add(id refs.Feed, until time.Time) error

	List() ([]refs.Feed, error)

	Contains(id refs.Feed) (bool, error)
}

type CurrentTimeProvider interface {
	Get() time.Time
}

// BannableRef wraps a feed ref.
type BannableRef struct {
	v any
}

func NewBannableRef(v any) (BannableRef, error) {
	switch v.(type) {
	case refs.Feed:
	default:
		return BannableRef{}, errors.New("must carry a feed ref")
	}
	return BannableRef{v: v}, nil
}

func (b *BannableRef) Value() any {
	return b.v
}

type BanListRepository interface {
	// Add adds a hash to the ban list.
	Add(hash bans.Hash) error

	// Remove removes a hash from the ban list. If a hash isn't in the ban list
	// no errors are returned.
	Remove(hash bans.Hash) error

	// ContainsFeed checks if the particular feed is banned.
	ContainsFeed(feed refs.Feed) (bool, error)

	// LookupMapping returns ErrBanListMappingNotFound error if a ref can not be
	// found.
	LookupMapping(hash bans.Hash) (BannableRef, error)
}

var ErrBanListMappingNotFound = errors.New("ban list mapping not found")

type Dialer interface {
	DialWithInitializer(ctx context.Context, initializer network.ClientPeerInitializer, remote identity.Public, addr network.Address) (transport.Peer, error)
	Dial(ctx context.Context, remote identity.Public, address network.Address) (transport.Peer, error)
}
