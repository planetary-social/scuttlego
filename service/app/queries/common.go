package queries

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type LogMessage struct {
	Message  message.Message
	Sequence ReceiveLogSequence
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

// ReceiveLogSequence is zero-indexed. This type has nothing to do with the
// sequence field of Scuttlebutt messages. It is a part of the system which
// simulates the behaviour of go-ssb's receive log.
type ReceiveLogSequence struct {
	seq int
}

func NewReceiveLogSequence(seq int) (ReceiveLogSequence, error) {
	if seq < 0 {
		return ReceiveLogSequence{}, errors.New("sequence can't be negative")
	}

	return ReceiveLogSequence{seq: seq}, nil
}

func MustNewReceiveLogSequence(seq int) ReceiveLogSequence {
	v, err := NewReceiveLogSequence(seq)
	if err != nil {
		panic(err)
	}

	return v
}

func (r ReceiveLogSequence) Int() int {
	return r.seq
}

type Dialer interface {
	Dial(ctx context.Context, remote identity.Public, address network.Address) (transport.Peer, error)
}
