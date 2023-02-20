package messagebuffer

import (
	"sort"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type FeedMessages struct {
	feed     refs.Feed
	messages []sortedMessage
}

func NewFeedMessages(feed refs.Feed) *FeedMessages {
	return &FeedMessages{feed: feed}
}

func (m *FeedMessages) Add(t time.Time, msg message.Message) error {
	if !msg.Feed().Equal(m.feed) {
		return errors.New("incorrect feed")
	}

	m.messages = append(m.messages, sortedMessage{
		Msg: msg,
		T:   t,
	})

	if l := len(m.messages); l >= 2 {
		last := l - 1
		if !m.messages[last].Msg.Sequence().ComesAfter(m.messages[last-1].Msg.Sequence()) {
			sort.Slice(m.messages, func(i, j int) bool {
				return m.messages[j].Msg.Sequence().ComesAfter(m.messages[i].Msg.Sequence())
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
		if !msg.Msg.Sequence().ComesAfter(sequence) {
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
func (m *FeedMessages) ConsecutiveSliceStartingWith(seq *message.Sequence) []message.Message {
	var result []message.Message

	for _, v := range m.messages {
		if seq != nil && !v.Msg.Sequence().ComesAfter(*seq) {
			continue
		}

		if l := len(result); l == 0 {
			if seq == nil {
				if !v.Msg.Sequence().IsFirst() {
					break
				}
			} else {
				if !seq.ComesDirectlyBefore(v.Msg.Sequence()) {
					break
				}
			}
		} else {
			if target := result[l-1].Sequence(); !target.ComesDirectlyBefore(v.Msg.Sequence()) && target != v.Msg.Sequence() {
				break
			}
		}

		result = append(result, v.Msg)
	}

	return result
}

func (m *FeedMessages) Sequences() []message.Sequence {
	var sequences []message.Sequence
	for _, msg := range m.messages {
		sequences = append(sequences, msg.Msg.Sequence())
	}
	return sequences
}

func (m *FeedMessages) Feed() refs.Feed {
	return m.feed
}

func (m *FeedMessages) Remove(msgToRemove message.Message) {
	for i, msg := range m.messages {
		if msg.Msg.Id().Equal(msgToRemove.Id()) {
			m.messages = append(m.messages[:i], m.messages[i+1:]...)
			return
		}
	}
}

type sortedMessage struct {
	Msg message.Message
	T   time.Time
}
