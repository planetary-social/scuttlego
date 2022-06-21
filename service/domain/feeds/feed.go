package feeds

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/blobs"
	msgcontents "github.com/planetary-social/go-ssb/service/domain/feeds/content"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

type Feed struct {
	lastMsg *message.Message

	format FeedFormat

	messagesToSave []message.Message
	contactsToSave []ContactToSave
	pubsToSave     []PubToSave
	blobsToSave    []BlobsToSave
}

func NewFeed(format FeedFormat) *Feed {
	return &Feed{
		format: format,
	}
}

func NewFeedFromHistory(lastMsg message.Message, format FeedFormat) (*Feed, error) {
	return &Feed{
		lastMsg: &lastMsg,
		format:  format,
	}, nil
}

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

func (f *Feed) CreateMessage(content message.RawMessageContent, timestamp time.Time, private identity.Private) (refs.Message, error) {
	if content.IsZero() {
		return refs.Message{}, errors.New("zero value of raw message content")
	}

	if timestamp.IsZero() {
		return refs.Message{}, errors.New("zero value of timestamp")
	}

	// todo check that private matches this feed's identity

	unsigned, err := f.createMessage(content, timestamp, private.Public())
	if err != nil {
		return refs.Message{}, errors.Wrap(err, "failed to create a new unsigned message")
	}

	msg, err := f.format.Sign(unsigned, private)
	if err != nil {
		return refs.Message{}, errors.Wrap(err, "failed to sign the new message")
	}

	err = f.onNewMessage(msg)
	if err != nil {
		return refs.Message{}, errors.Wrap(err, "failed to sign the new message")
	}

	return msg.Id(), nil
}

func (f *Feed) createMessage(content message.RawMessageContent, timestamp time.Time, author identity.Public) (message.UnsignedMessage, error) {
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
			message.NewFirstSequence(),
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

func (f *Feed) PopForPersisting() ([]message.Message, []ContactToSave, []PubToSave, []BlobsToSave) {
	defer func() { f.messagesToSave = nil; f.contactsToSave = nil; f.pubsToSave = nil }()
	return f.messagesToSave, f.contactsToSave, f.pubsToSave, f.blobsToSave
}

func (f *Feed) onNewMessage(msg message.Message) error {
	contacts, pubs, blobs := f.processMessageContent(msg)

	f.lastMsg = &msg
	f.messagesToSave = append(f.messagesToSave, msg)
	f.contactsToSave = append(f.contactsToSave, contacts...)
	f.pubsToSave = append(f.pubsToSave, pubs...)
	f.blobsToSave = append(f.blobsToSave, blobs...)
	return nil
}

// todo ignore repeated calls to follow someone if the current state of the feed suggests that this is already done (indempotency)
func (f *Feed) processMessageContent(msg message.Message) ([]ContactToSave, []PubToSave, []BlobsToSave) {
	return f.getContactsToSave(msg), f.getPubsToSave(msg), f.getBlobsToSave(msg)
}

func (f *Feed) getContactsToSave(msg message.Message) []ContactToSave {
	switch v := msg.Content().(type) {
	case msgcontents.Contact:
		return []ContactToSave{NewContactToSave(msg.Author(), v)}
	default:
		return nil
	}
}

func (f *Feed) getPubsToSave(msg message.Message) []PubToSave {
	switch v := msg.Content().(type) {
	case msgcontents.Pub:
		return []PubToSave{NewPubToSave(msg.Author(), msg.Id(), v)}
	default:
		return nil
	}
}

func (f *Feed) getBlobsToSave(msg message.Message) []BlobsToSave {
	blobReferencer, ok := msg.Content().(blobs.BlobReferencer)
	if !ok {
		return nil
	}
	if blobRefs := blobReferencer.Blobs(); len(blobRefs) > 0 {
		return []BlobsToSave{NewBlobsToSave(msg.Feed(), msg.Id(), blobRefs)}
	}
	return nil
}
