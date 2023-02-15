package feeds

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type MessageToPersist struct {
	msg            message.Message
	contactsToSave []ContactToSave
	pubsToSave     []PubToSave
	blobsToSave    []BlobToSave
}

func NewMessageToPersist(
	msg message.Message,
	contactsToSave []ContactToSave,
	pubsToSave []PubToSave,
	blobsToSave []BlobToSave,
) (MessageToPersist, error) {
	if msg.IsZero() {
		return MessageToPersist{}, errors.New("zero value of message")
	}

	return MessageToPersist{
		msg:            msg,
		contactsToSave: contactsToSave,
		pubsToSave:     pubsToSave,
		blobsToSave:    blobsToSave,
	}, nil
}

func MustNewMessageToPersist(
	msg message.Message,
	contactsToSave []ContactToSave,
	pubsToSave []PubToSave,
	blobsToSave []BlobToSave,
) MessageToPersist {
	v, err := NewMessageToPersist(msg, contactsToSave, pubsToSave, blobsToSave)
	if err != nil {
		panic(err)
	}
	return v
}

func (m MessageToPersist) Message() message.Message {
	return m.msg
}

func (m MessageToPersist) ContactsToSave() []ContactToSave {
	return m.contactsToSave
}

func (m MessageToPersist) PubsToSave() []PubToSave {
	return m.pubsToSave
}

func (m MessageToPersist) BlobsToSave() []BlobToSave {
	return m.blobsToSave
}

type PubToSave struct {
	who     refs.Identity
	message refs.Message
	content known.Pub
}

func NewPubToSave(who refs.Identity, message refs.Message, content known.Pub) PubToSave {
	return PubToSave{
		who:     who,
		message: message,
		content: content,
	}
}

func (c PubToSave) Who() refs.Identity {
	return c.who
}

func (c PubToSave) Message() refs.Message {
	return c.message
}

func (c PubToSave) Content() known.Pub {
	return c.content
}

type BlobToSave struct {
	ref refs.Blob
}

func NewBlobToSave(ref refs.Blob) (BlobToSave, error) {
	if ref.IsZero() {
		return BlobToSave{}, errors.New("zero value of ref")
	}

	return BlobToSave{
		ref: ref,
	}, nil
}

func MustNewBlobToSave(ref refs.Blob) BlobToSave {
	v, err := NewBlobToSave(ref)
	if err != nil {
		panic(err)
	}
	return v
}

func (b BlobToSave) Ref() refs.Blob {
	return b.ref
}

type ContactToSave struct {
	who refs.Identity
	msg known.Contact
}

func NewContactToSave(who refs.Identity, msg known.Contact) ContactToSave {
	return ContactToSave{
		who: who,
		msg: msg,
	}
}

func (c ContactToSave) Who() refs.Identity {
	return c.who
}

func (c ContactToSave) Msg() known.Contact {
	return c.msg
}
