package feeds

import (
	"time"

	"github.com/boreq/errors"
	msgcontents "github.com/planetary-social/go-ssb/service/domain/feeds/content"
	message2 "github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

type Feed struct {
	lastMsg *message2.Message

	format FeedFormat

	messagesToSave []message2.Message
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

func NewFeedFromHistory(lastMsg message2.Message, format FeedFormat) (*Feed, error) {
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

func (f *Feed) AppendMessage(msg message2.Message) error {
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

func (f *Feed) CreateMessage(content message2.MessageContent, timestamp time.Time, private identity.Private) error {
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

func (f *Feed) createMessage(content message2.MessageContent, timestamp time.Time, author identity.Public) (message2.UnsignedMessage, error) {
	authorRef, err := refs.NewIdentityFromPublic(author)
	if err != nil {
		return message2.UnsignedMessage{}, errors.Wrap(err, "could not create an author")
	}

	if f.lastMsg != nil {
		previousId := f.lastMsg.Id()
		return message2.NewUnsignedMessage(
			&previousId,
			f.lastMsg.Sequence().Next(),
			authorRef,
			f.lastMsg.Feed(),
			timestamp,
			content,
		)
	} else {
		return message2.NewUnsignedMessage(
			nil,
			message2.FirstSequence,
			authorRef,
			authorRef.MainFeed(),
			timestamp,
			content,
		)
	}
}

func (f *Feed) Sequence() message2.Sequence {
	return f.lastMsg.Sequence() // todo can be nil
}

func (f *Feed) PopForPersisting() ([]message2.Message, []ContactToSave) {
	defer func() { f.messagesToSave = nil; f.contactsToSave = nil }()
	return f.messagesToSave, f.contactsToSave
}

func (f *Feed) onNewMessage(msg message2.Message) error {
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
func (f *Feed) processMessageContent(msg message2.Message) ([]ContactToSave, error) {
	switch v := msg.Content().(type) {
	case msgcontents.Contact:
		return []ContactToSave{NewContactToSave(msg.Author(), v)}, nil
	default:
		return nil, nil
	}
}
