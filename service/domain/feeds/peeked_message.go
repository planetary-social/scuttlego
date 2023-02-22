package feeds

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type PeekedMessage struct {
	feed     refs.Feed
	sequence message.Sequence
	raw      message.RawMessage
}

func NewPeekedMessage(
	feed refs.Feed,
	sequence message.Sequence,
	raw message.RawMessage,
) (PeekedMessage, error) {
	if feed.IsZero() {
		return PeekedMessage{}, errors.New("zero value of feed")
	}
	if sequence.IsZero() {
		return PeekedMessage{}, errors.New("zero value of sequence")
	}
	if raw.IsZero() {
		return PeekedMessage{}, errors.New("zero value of raw")
	}
	return PeekedMessage{
		feed:     feed,
		sequence: sequence,
		raw:      raw,
	}, nil
}

func MustNewPeekedMessage(
	feed refs.Feed,
	sequence message.Sequence,
	raw message.RawMessage,
) PeekedMessage {
	v, err := NewPeekedMessage(feed, sequence, raw)
	if err != nil {
		panic(err)
	}
	return v
}

func (p PeekedMessage) Feed() refs.Feed {
	return p.feed
}

func (p PeekedMessage) Sequence() message.Sequence {
	return p.sequence
}

func (p PeekedMessage) Raw() message.RawMessage {
	return p.raw
}

func (p PeekedMessage) IsZero() bool {
	return p.feed.IsZero()
}
