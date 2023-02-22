package messagebuffer

import (
	"bytes"
	"sort"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type IsZeroer interface {
	IsZero() bool
}

type ReceivedMessage[T IsZeroer] struct {
	replicatedFrom identity.Public
	message        T
}

func NewReceivedMessage[T IsZeroer](replicatedFrom identity.Public, message T) (ReceivedMessage[T], error) {
	if replicatedFrom.IsZero() {
		return ReceivedMessage[T]{}, errors.New("zero value of replicated from")
	}
	if message.IsZero() {
		return ReceivedMessage[T]{}, errors.New("zero value of message")
	}
	return ReceivedMessage[T]{replicatedFrom: replicatedFrom, message: message}, nil
}

func MustNewReceivedMessage[T IsZeroer](replicatedFrom identity.Public, message T) ReceivedMessage[T] {
	v, err := NewReceivedMessage(replicatedFrom, message)
	if err != nil {
		panic(err)
	}
	return v
}

func (r *ReceivedMessage[T]) ReplicatedFrom() identity.Public {
	return r.replicatedFrom
}

func (r *ReceivedMessage[T]) Message() T {
	return r.message
}

type FeedMessages struct {
	feed     refs.Feed
	messages []sortedMessage
}

func NewFeedMessages(feed refs.Feed) *FeedMessages {
	return &FeedMessages{feed: feed}
}

func (m *FeedMessages) Add(t time.Time, rm ReceivedMessage[feeds.PeekedMessage]) error {
	if !rm.Message().Feed().Equal(m.feed) {
		return errors.New("incorrect feed")
	}

	m.messages = append(m.messages, sortedMessage{
		ReceivedMessage: rm,
		T:               t,
	})

	if l := len(m.messages); l >= 2 {
		last := l - 1
		if !m.messages[last].Message().Sequence().ComesAfter(m.messages[last-1].Message().Sequence()) {
			sort.Slice(m.messages, func(i, j int) bool {
				return m.messages[j].Message().Sequence().ComesAfter(m.messages[i].Message().Sequence())
			})
		}
	}

	return nil
}

func (m *FeedMessages) RemoveOlderThan(t time.Time) {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].T.Before(t) {
			m.messages = append(m.messages[:i], m.messages[i+1:]...)
		}
	}
}

func (m *FeedMessages) LeaveOnlyAfter(sequence message.Sequence) {
	var foundIndex *int
	for i, msg := range m.messages {
		if !msg.Message().Sequence().ComesAfter(sequence) {
			tmp := i
			foundIndex = &tmp
		} else {
			break
		}
	}

	if foundIndex != nil {
		m.messages = m.messages[*foundIndex+1:]
	}
}

func (m *FeedMessages) Len() int {
	return len(m.messages)
}

// ConsecutiveSliceStartingWith returns a slice of messages that have
// consecutive sequence numbers. The first message in the slice has a sequence
// number which comes directly after the provided sequence number. If the
// provided sequence number is nil then the first element has a sequence number
// equal to the output of message.NewFirstSequence. If those conditions can't be
// satisfied an empty slice is returned.
func (m *FeedMessages) ConsecutiveSliceStartingWith(seq *message.Sequence) []ReceivedMessage[feeds.PeekedMessage] {
	var result []ReceivedMessage[feeds.PeekedMessage]

	for _, v := range m.messages {
		if seq != nil && !v.Message().Sequence().ComesAfter(*seq) {
			continue
		}

		if l := len(result); l == 0 {
			if seq == nil {
				if !v.Message().Sequence().IsFirst() {
					break
				}
			} else {
				if !seq.ComesDirectlyBefore(v.Message().Sequence()) {
					break
				}
			}
		} else {
			if target := result[l-1].Message().Sequence(); !target.ComesDirectlyBefore(v.Message().Sequence()) && target != v.Message().Sequence() {
				break
			}
		}

		result = append(result, v.ReceivedMessage)
	}

	return result
}

func (m *FeedMessages) Sequences() []message.Sequence {
	var sequences []message.Sequence
	for _, msg := range m.messages {
		sequences = append(sequences, msg.Message().Sequence())
	}
	return sequences
}

func (m *FeedMessages) Feed() refs.Feed {
	return m.feed
}

func (m *FeedMessages) Remove(msgToRemove message.RawMessage) {
	for i, msg := range m.messages {
		if bytes.Equal(msg.Message().Raw().Bytes(), msgToRemove.Bytes()) {
			m.messages = append(m.messages[:i], m.messages[i+1:]...)
			return
		}
	}
}

type sortedMessage struct {
	ReceivedMessage[feeds.PeekedMessage]
	T time.Time
}
