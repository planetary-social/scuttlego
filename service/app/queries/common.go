package queries

import (
	"context"

	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type LogMessage struct {
	Message  message.Message
	Sequence common.ReceiveLogSequence
}

type Dialer interface {
	Dial(ctx context.Context, remote identity.Public, address network.Address) (transport.Peer, error)
}

type TransactionProvider interface {
	Transact(func(adapters Adapters) error) error
}

type Adapters struct {
	Feed         FeedRepository
	ReceiveLog   ReceiveLogRepository
	Message      MessageRepository
	SocialGraph  SocialGraphRepository
	FeedWantList FeedWantListRepository
	BanList      BanListRepository
}
type FeedRepository interface {
	// GetMessages returns messages with a sequence greater or equal to the
	// provided sequence. If sequence is nil then messages starting from the
	// beginning of the feed are returned. Limit specifies the max number of
	// returned messages. If limit is nil then all messages matching the
	// sequence criteria are returned.
	GetMessages(id refs.Feed, seq *message.Sequence, limit *int) ([]message.Message, error)

	// GetFeed returns a feed so that you can for example check its sequence
	// number. Returns common.ErrFeedNotFound if the feed doesn't exist.
	GetFeed(ref refs.Feed) (*feeds.Feed, error)

	// GetMessage returns a message with a given sequence from the specified
	// feed.
	GetMessage(feed refs.Feed, sequence message.Sequence) (message.Message, error)

	// Count returns the number of stored feeds.
	Count() (int, error)
}

type ReceiveLogRepository interface {
	// List returns messages from the log starting with the provided sequence.
	// This is supposed to simulate the behaviour of go-ssb's receive log as
	// such a concept doesn't exist within this implementation. The log is zero
	// indexed. If limit isn't positive an error is returned. Sequence has
	// nothing to do with the sequence field of Scuttlebutt messages.
	List(startSeq common.ReceiveLogSequence, limit int) ([]LogMessage, error)

	// GetMessage returns the message that the provided receive log sequence
	// points to.
	GetMessage(seq common.ReceiveLogSequence) (message.Message, error)

	// GetSequences returns the sequences assigned to a message in the receive
	// log. If an error isn't returned then the slice will have at least one
	// element.
	GetSequences(ref refs.Message) ([]common.ReceiveLogSequence, error)
}

type MessageRepository interface {
	// Count returns the number of stored messages.
	Count() (int, error)

	// Get retrieves a message.
	Get(id refs.Message) (message.Message, error)
}

type SocialGraphRepository interface {
	GetSocialGraph() (graph.SocialGraph, error)
}

type FeedWantListRepository interface {
	List() ([]refs.Feed, error)
}

type BanListRepository interface {
	ContainsFeed(feed refs.Feed) (bool, error)
}
