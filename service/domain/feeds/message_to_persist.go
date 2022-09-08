package feeds

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
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
	content content.Pub
}

func NewPubToSave(who refs.Identity, message refs.Message, content content.Pub) PubToSave {
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

func (c PubToSave) Content() content.Pub {
	return c.content
}

type BlobToSave struct {
	blobs []refs.Blob
}

func NewBlobToSave(blobs []refs.Blob) BlobToSave {
	return BlobToSave{
		blobs: blobs,
	}
}

func (b BlobToSave) Blobs() []refs.Blob {
	return b.blobs
}

type ContactToSave struct {
	who refs.Identity
	msg content.Contact
}

func NewContactToSave(who refs.Identity, msg content.Contact) ContactToSave {
	return ContactToSave{
		who: who,
		msg: msg,
	}
}

func (c ContactToSave) Who() refs.Identity {
	return c.who
}

func (c ContactToSave) Msg() content.Contact {
	return c.msg
}
