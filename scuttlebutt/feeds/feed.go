package feeds

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/identity"
	"github.com/planetary-social/go-ssb/refs"
	msgcontents "github.com/planetary-social/go-ssb/scuttlebutt/feeds/content"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
)

type Feed struct {
	lastMsg *message.Message

	format FeedFormat

	messagesToSave []message.Message
	contactsToSave []ContactToSave
}

func NewFeed(format FeedFormat) *Feed {
	return &Feed{
		format: format,
	}
}

//func NewFeedFromMessageContent(content message.MessageContent, timestamp time.Time, private identity.Private) (*Feed, error) {
//	// todo identify feed format, right now lets just hardcode it as we support only default ones
//	//format := formats.NewScuttlebutt()
//
//	//author, err := refs.NewIdentityFromPublic(private.Public())
//	//if err != nil {
//	//	return nil, errors.Wrap(err, "could not create an author")
//	//}
//
//	//unsigned, err := message.NewUnsignedMessage(
//	//	nil,
//	//	message.FirstSequence,
//	//	author,
//	//	author.MainFeed(),
//	//	timestamp,
//	//	content,
//	//)
//	//if err != nil {
//	//	return nil, errors.Wrap(err, "failed to create a new unsigned message")
//	//}
//
//	//msg, err := format.Sign(unsigned, private)
//	//if err != nil {
//	//	return nil, errors.Wrap(err, "failed to sign the new message")
//	//}
//
//	//return newFeedFromMessage(msg)
//	return nil, errors.New("not implemented")
//}

func NewFeedFromHistory(lastMsg message.Message, format FeedFormat) (*Feed, error) {
	return &Feed{
		lastMsg: &lastMsg,
		format:  format,
	}, nil
}

//func newFeedFromMessage(msg message.Message) (*Feed, error) {
//	f := &Feed{}
//	if err := f.onNewMessage(msg); err != nil {
//		return nil, errors.Wrap(err, "failed to process a new message")
//	}
//	return f, nil
//}

func (f *Feed) AppendMessage(msg message.Message) error {
	if f.lastMsg != nil {
		if !msg.Sequence().ComesAfter(f.lastMsg.Sequence()) {
			return nil // idempotency
		}

		if msg.IsZero() {
			return errors.New("zero value of message")
		}

		if !msg.Author().Equal(f.lastMsg.Author()) {
			return errors.New("invalid author")
		}

		if !msg.Feed().Equal(f.lastMsg.Feed()) {
			return errors.New("invalid feed")
		}

		if msg.IsRootMessage() {
			return errors.New("can't append the root message")
		}

		if !f.lastMsg.ComesDirectlyBefore(msg) {
			return errors.New("this is not the next message in this feed")
		}
	} else {
		if !msg.IsRootMessage() {
			return errors.New("first message in the feed must be a root message")
		}
	}

	return f.onNewMessage(msg)
}

func (f *Feed) CreateMessage(content message.MessageContent, timestamp time.Time, private identity.Private) error {
	unsigned, err := f.createMessage(content, timestamp, private.Public())
	if err != nil {
		return errors.Wrap(err, "failed to create a new unsigned message")
	}

	msg, err := f.format.Sign(unsigned, private)
	if err != nil {
		return errors.Wrap(err, "failed to sign the new message")
	}

	return f.onNewMessage(msg)
}

func (f *Feed) createMessage(content message.MessageContent, timestamp time.Time, author identity.Public) (message.UnsignedMessage, error) {
	authorRef, err := refs.NewIdentityFromPublic(author)
	if err != nil {
		return message.UnsignedMessage{}, errors.Wrap(err, "could not create an author")
	}

	if f.lastMsg != nil {
		previousId := f.lastMsg.Id()
		return message.NewUnsignedMessage(
			&previousId,
			f.lastMsg.Sequence().Next(),
			authorRef,
			f.lastMsg.Feed(),
			timestamp,
			content,
		)
	} else {
		return message.NewUnsignedMessage(
			nil,
			message.FirstSequence,
			authorRef,
			authorRef.MainFeed(),
			timestamp,
			content,
		)
	}
}

func (f *Feed) Sequence() message.Sequence {
	return f.lastMsg.Sequence() // todo can be nil
}

func (f *Feed) PopForPersisting() ([]message.Message, []ContactToSave) {
	defer func() { f.messagesToSave = nil; f.contactsToSave = nil }()
	return f.messagesToSave, f.contactsToSave
}

func (f *Feed) onNewMessage(msg message.Message) error {
	contacts, err := f.processMessageContent(msg)
	if err != nil {
		return errors.New("failed to process message content")
	}

	f.lastMsg = &msg
	f.messagesToSave = append(f.messagesToSave, msg)
	f.contactsToSave = append(f.contactsToSave, contacts...)
	return nil
}

// todo ignore repeated calls to follow someone if the current state of the feed suggests that this is already done (indempotency)
func (f *Feed) processMessageContent(msg message.Message) ([]ContactToSave, error) {
	switch v := msg.Content().(type) {
	case msgcontents.Contact:
		return []ContactToSave{NewContactToSave(msg.Author(), v)}, nil
	default:
		return nil, nil
	}
}
