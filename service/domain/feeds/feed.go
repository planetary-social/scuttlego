package feeds

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	msgcontents "github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type Feed struct {
	lastMsg        *message.Message
	format         FeedFormat
	messagesToSave []MessageToPersist
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

	if private.IsZero() {
		return refs.Message{}, errors.New("zero value of private identity")
	}

	if f.lastMsg != nil && !f.lastMsg.Author().Identity().Equal(private.Public()) {
		return refs.Message{}, errors.New("private identity doesn't match this feed's public identity")
	}

	unsigned, err := f.createMessage(content, timestamp, private.Public())
	if err != nil {
		return refs.Message{}, errors.Wrap(err, "failed to create a new unsigned message")
	}

	msg, err := f.format.Sign(unsigned, private)
	if err != nil {
		return refs.Message{}, errors.Wrap(err, "failed to sign the new message")
	}

	if err := f.AppendMessage(msg); err != nil {
		return refs.Message{}, errors.Wrap(err, "failed to append the new message")
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

func (f *Feed) Sequence() (message.Sequence, bool) {
	if f.lastMsg == nil {
		return message.Sequence{}, false
	}
	return f.lastMsg.Sequence(), true
}

func (f *Feed) PopForPersisting() []MessageToPersist {
	defer func() { f.messagesToSave = nil }()
	return f.messagesToSave
}

// todo ignore repeated calls to follow someone if the current state of the feed suggests that this is already done (indempotency)
func (f *Feed) onNewMessage(msg message.Message) error {
	contacts := f.getContactsToSave(msg)
	pubs := f.getPubsToSave(msg)
	blobs := f.getBlobsToSave(msg)

	msgToSave, err := NewMessageToPersist(msg, contacts, pubs, blobs)
	if err != nil {
		return errors.Wrap(err, "failed to create a message to save")
	}

	f.lastMsg = &msg
	f.messagesToSave = append(f.messagesToSave, msgToSave)
	return nil
}

func (f *Feed) getContactsToSave(msg message.Message) []ContactToSave {
	switch v := msg.Content().(type) {
	case msgcontents.Contact:
		return []ContactToSave{NewContactToSave(msg.Author(), v)} // todo author or feed
	default:
		return nil
	}
}

func (f *Feed) getPubsToSave(msg message.Message) []PubToSave {
	switch v := msg.Content().(type) {
	case msgcontents.Pub:
		return []PubToSave{NewPubToSave(msg.Author(), msg.Id(), v)} // todo author or feed
	default:
		return nil
	}
}

func (f *Feed) getBlobsToSave(msg message.Message) []BlobToSave {
	blobReferencer, ok := msg.Content().(blobs.BlobReferencer)
	if !ok {
		return nil
	}
	if blobRefs := blobReferencer.Blobs(); len(blobRefs) > 0 {
		return []BlobToSave{NewBlobToSave(blobRefs)}
	}
	return nil
}
