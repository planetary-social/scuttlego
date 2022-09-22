package replication

import (
	"strconv"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
)

type RawMessageHandler interface {
	Handle(msg message.RawMessage) error
}

// FeedState wraps the sequence number so that both the state of feeds which
// have some messages in them and empty feeds can be represented.
type FeedState struct {
	sequence *message.Sequence
}

// NewEmptyFeedState creates a new feed state which represents an empty feed.
// This is equivalent to the zero value of this type but using this constructor
// improves readability.
func NewEmptyFeedState() FeedState {
	return FeedState{}
}

// NewFeedState creates a new feed state which represents a feed for which at
// least one message is known.
func NewFeedState(sequence message.Sequence) (FeedState, error) {
	if sequence.IsZero() {
		return FeedState{}, errors.New("zero value of sequence")
	}

	return FeedState{
		sequence: &sequence,
	}, nil
}

func MustNewFeedState(sequence message.Sequence) FeedState {
	v, err := NewFeedState(sequence)
	if err != nil {
		panic(err)
	}
	return v
}

// Sequence returns the sequence of the last message in the feed. If the feed is
// empty then the sequence is not returned.
func (s FeedState) Sequence() (message.Sequence, bool) {
	if s.sequence != nil {
		return *s.sequence, true
	}
	return message.Sequence{}, false
}

// String is useful for printing this value when logging or debugging. Do not
// use it for other purposes.
func (s FeedState) String() string {
	if s.sequence != nil {
		return strconv.Itoa(s.sequence.Int())
	}
	return "empty"
}
