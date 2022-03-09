package message

import (
	refs2 "github.com/planetary-social/go-ssb/service/domain/refs"
	"time"

	"github.com/boreq/errors"
)

type RawMessage struct {
	data []byte
}

func NewRawMessage(data []byte) RawMessage {
	return RawMessage{
		data: data,
	}
}

func (m RawMessage) Bytes() []byte {
	tmp := make([]byte, len(m.data))
	copy(tmp, m.data)
	return tmp
}

func (m RawMessage) IsZero() bool {
	return len(m.data) == 0
}

type UnsignedMessage struct {
	baseMessageFields
}

func NewUnsignedMessage(
	previous *refs2.Message,
	sequence Sequence,
	author refs2.Identity,
	feed refs2.Feed,
	timestamp time.Time,
	content MessageContent,
) (UnsignedMessage, error) {
	fields, err := newBaseMessageFields(previous, sequence, author, feed, timestamp, content)
	if err != nil {
		return UnsignedMessage{}, errors.New("could not create known message fields")
	}

	return UnsignedMessage{
		baseMessageFields: fields,
	}, nil
}

type Message struct {
	baseMessageFields

	id  refs2.Message
	raw RawMessage
}

func NewMessage(
	id refs2.Message,
	previous *refs2.Message,
	sequence Sequence,
	author refs2.Identity,
	feed refs2.Feed,
	timestamp time.Time,
	content MessageContent,
	raw RawMessage,
) (Message, error) {
	fields, err := newBaseMessageFields(previous, sequence, author, feed, timestamp, content)
	if err != nil {
		return Message{}, errors.New("could not create known message fields")
	}

	if id.IsZero() {
		return Message{}, errors.New("zero value of id")
	}

	if raw.IsZero() {
		return Message{}, errors.New("zero value of raw message")
	}

	return Message{
		id:                id,
		baseMessageFields: fields,
		raw:               raw,
	}, nil
}

func MustNewMessage(
	id refs2.Message,
	previous *refs2.Message,
	sequence Sequence,
	author refs2.Identity,
	feed refs2.Feed,
	timestamp time.Time,
	content MessageContent,
	raw RawMessage,
) Message {
	msg, err := NewMessage(id, previous, sequence, author, feed, timestamp, content, raw)
	if err != nil {
		panic(err)
	}
	return msg
}

func (m Message) Id() refs2.Message {
	return m.id
}

func (m Message) Raw() RawMessage {
	return m.raw
}

func (m Message) IsRootMessage() bool {
	return m.sequence.IsFirst()
}

func (m Message) ComesDirectlyBefore(o Message) bool {
	if o.previous == nil {
		return false
	}
	return m.sequence.ComesDirectlyBefore(o.sequence) && o.previous.Equal(m.id)
}

func (m Message) IsZero() bool {
	return m.id.IsZero()
}

// One of:
// - Contact
// - Pub
// - Unknown
// todo generics once they are available
type MessageContent interface {
	Type() MessageContentType
}

type MessageContentType string // todo struct with strings.ToLower or pragma nocompare?

func (t MessageContentType) IsZero() bool {
	return t == ""
}

type baseMessageFields struct {
	previous  *refs2.Message
	sequence  Sequence
	author    refs2.Identity
	feed      refs2.Feed
	timestamp time.Time
	content   MessageContent
}

func newBaseMessageFields(
	previous *refs2.Message,
	sequence Sequence,
	author refs2.Identity,
	feed refs2.Feed,
	timestamp time.Time,
	content MessageContent,
) (baseMessageFields, error) {
	if previous != nil && previous.IsZero() {
		return baseMessageFields{}, errors.New("zero value of previous")
	}

	if previous != nil && sequence.IsFirst() {
		return baseMessageFields{}, errors.New("this message has a previous message so it can't be first")
	}

	if previous == nil && !sequence.IsFirst() {
		return baseMessageFields{}, errors.New("this message doesn't have a previous message so it must be first")
	}

	if author.IsZero() {
		return baseMessageFields{}, errors.New("zero value of author")
	}

	if feed.IsZero() {
		return baseMessageFields{}, errors.New("zero value of feed")
	}

	// there is no way to validate the timestamp as it can be set to anything

	if content == nil {
		return baseMessageFields{}, errors.New("nil content")
	}

	return baseMessageFields{
		previous:  previous,
		sequence:  sequence,
		author:    author,
		feed:      feed,
		timestamp: timestamp,
		content:   content,
	}, nil
}

func (k baseMessageFields) Previous() *refs2.Message {
	return k.previous
}

func (k baseMessageFields) Sequence() Sequence {
	return k.sequence
}

func (k baseMessageFields) Author() refs2.Identity {
	return k.author
}

func (k baseMessageFields) Feed() refs2.Feed {
	return k.feed
}

func (k baseMessageFields) Timestamp() time.Time {
	return k.timestamp
}

func (k baseMessageFields) Content() MessageContent {
	return k.content
}
