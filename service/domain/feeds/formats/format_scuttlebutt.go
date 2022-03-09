package formats

import (
	"encoding/json"
	message2 "github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	refs2 "github.com/planetary-social/go-ssb/service/domain/refs"
	"time"

	"github.com/boreq/errors"
	"go.cryptoscope.co/ssb/message/legacy"
	ssbrefs "go.mindeco.de/ssb-refs"
)

type Scuttlebutt struct {
	marshaler Marshaler
}

func NewScuttlebutt(marshaler Marshaler) *Scuttlebutt {
	return &Scuttlebutt{
		marshaler: marshaler,
	}
}

// todo should we have a way to set hmac for testing purposes? what this hmac is for:
// https://github.com/ssb-js/ssb-validate#state--validateappendstate-hmac_key-msg

func (s *Scuttlebutt) Verify(raw message2.RawMessage) (message2.Message, error) {
	ssbRef, ssbMessage, err := legacy.Verify(raw.Bytes(), nil)
	if err != nil {
		return message2.Message{}, errors.Wrap(err, "verification failed")
	}

	id, err := refs2.NewMessage(ssbRef.Ref())
	if err != nil {
		return message2.Message{}, errors.Wrap(err, "invalid message id")
	}

	var previous *refs2.Message
	if ssbMessage.Previous != nil {
		tmp, err := refs2.NewMessage(ssbMessage.Previous.Ref())
		if err != nil {
			return message2.Message{}, errors.Wrap(err, "invalid previous message id")
		}
		previous = &tmp
	}

	sequence, err := message2.NewSequence(int(ssbMessage.Sequence))
	if err != nil {
		return message2.Message{}, errors.Wrap(err, "invalid sequence")
	}

	author, err := refs2.NewIdentity(ssbMessage.Author.Ref())
	if err != nil {
		return message2.Message{}, errors.Wrap(err, "invalid author")
	}

	feed, err := refs2.NewFeed(ssbMessage.Author.Ref())
	if err != nil {
		return message2.Message{}, errors.Wrap(err, "invalid feed")
	}

	timestamp := time.UnixMilli(int64(ssbMessage.Timestamp))

	content, err := s.marshaler.Unmarshal(ssbMessage.Content)
	if err != nil {
		return message2.Message{}, errors.Wrap(err, "could not unmarshal message content")
	}

	msg, err := message2.NewMessage(
		id,
		previous,
		sequence,
		author,
		feed,
		timestamp,
		content,
		raw,
	)
	if err != nil {
		return message2.Message{}, errors.Wrap(err, "could not create a message")
	}

	return msg, nil
}

func (s *Scuttlebutt) Sign(unsigned message2.UnsignedMessage, private identity.Private) (message2.Message, error) {
	var previous *ssbrefs.MessageRef
	if unsigned.Previous() != nil {
		tmp, err := ssbrefs.NewMessageRefFromBytes(unsigned.Previous().Bytes(), ssbrefs.RefAlgoMessageSSB1)
		if err != nil {
			return message2.Message{}, errors.New("could not create a ref")
		}
		previous = &tmp
	}

	marshaledContent, err := s.marshaler.Marshal(unsigned.Content())
	if err != nil {
		return message2.Message{}, errors.Wrap(err, "could not marshal message content")
	}

	msgToSign := legacy.LegacyMessage{
		Previous:  previous,
		Author:    unsigned.Author().String(),
		Sequence:  int64(unsigned.Sequence().Int()),
		Timestamp: unsigned.Timestamp().UnixMilli(),
		Hash:      "sha256",
		Content:   json.RawMessage(marshaledContent),
	}

	_, raw, err := msgToSign.Sign(private.PrivateKey(), nil)
	if err != nil {
		return message2.Message{}, errors.New("could not sign a message")
	}

	return s.Verify(message2.NewRawMessage(raw))
}
