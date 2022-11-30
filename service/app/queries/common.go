package queries

import (
	"context"

	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type LogMessage struct {
	Message  message.Message
	Sequence common.ReceiveLogSequence
}

type FeedRepository interface {
	// GetMessages returns messages with a sequence greater or equal to the
	// provided sequence. If sequence is nil then messages starting from the
	// beginning of the feed are returned. Limit specifies the max number of
	// returned messages. If limit is nil then all messages matching the
	// sequence criteria are returned.
	GetMessages(id refs.Feed, seq *message.Sequence, limit *int) ([]message.Message, error) // todo iterator instead of returning a huge array

	// GetMessage returns a message with a given sequence from the specified
	// feed.
	GetMessage(feed refs.Feed, sequence message.Sequence) (message.Message, error)

	// Count returns the number of stored feeds.
	Count() (int, error)
}

type Dialer interface {
	Dial(ctx context.Context, remote identity.Public, address network.Address) (transport.Peer, error)
}
