package formats

import (
	"encoding/json"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/identity"
	"github.com/planetary-social/go-ssb/refs"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
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

func (s *Scuttlebutt) Verify(raw message.RawMessage) (message.Message, error) {
	ssbRef, ssbMessage, err := legacy.Verify(raw.Bytes(), nil)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "verification failed")
	}

	id, err := refs.NewMessage(ssbRef.Ref())
	if err != nil {
		return message.Message{}, errors.Wrap(err, "invalid message id")
	}

	var previous *refs.Message
	if ssbMessage.Previous != nil {
		tmp, err := refs.NewMessage(ssbMessage.Previous.Ref())
		if err != nil {
			return message.Message{}, errors.Wrap(err, "invalid previous message id")
		}
		previous = &tmp
	}

	sequence, err := message.NewSequence(int(ssbMessage.Sequence))
	if err != nil {
		return message.Message{}, errors.Wrap(err, "invalid sequence")
	}

	author, err := refs.NewIdentity(ssbMessage.Author.Ref())
	if err != nil {
		return message.Message{}, errors.Wrap(err, "invalid author")
	}

	feed, err := refs.NewFeed(ssbMessage.Author.Ref())
	if err != nil {
		return message.Message{}, errors.Wrap(err, "invalid feed")
	}

	timestamp := time.UnixMilli(int64(ssbMessage.Timestamp))

	content, err := s.marshaler.Unmarshal(ssbMessage.Content)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "could not unmarshal message content")
	}

	msg, err := message.NewMessage(
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
		return message.Message{}, errors.Wrap(err, "could not create a message")
	}

	return msg, nil
}

func (s *Scuttlebutt) Sign(unsigned message.UnsignedMessage, private identity.Private) (message.Message, error) {
	var previous *ssbrefs.MessageRef
	if unsigned.Previous() != nil {
		tmp, err := ssbrefs.NewMessageRefFromBytes(unsigned.Previous().Bytes(), ssbrefs.RefAlgoFeedSSB1)
		if err != nil {
			return message.Message{}, errors.New("could not create a ref")
		}
		previous = &tmp
	}

	marshaledContent, err := s.marshaler.Marshal(unsigned.Content())
	if err != nil {
		return message.Message{}, errors.Wrap(err, "could not marshal message content")
	}

	msgToSign := legacy.LegacyMessage{
		Previous:  previous,
		Author:    unsigned.Author().String(),
		Sequence:  int64(unsigned.Sequence().Int()),
		Timestamp: unsigned.Timestamp().UnixMilli(),
		Hash:      "sha256",
		Content:   json.RawMessage(marshaledContent),
	}

	ssbRef, raw, err := msgToSign.Sign(private.PrivateKey(), nil)
	if err != nil {
		return message.Message{}, errors.New("could not sign a message")
	}

	id, err := refs.NewMessage(ssbRef.Ref())
	if err != nil {
		return message.Message{}, errors.Wrap(err, "invalid message id")
	}

	// todo maybe just return using a call to Verify?
	return s.Verify(message.NewRawMessage(raw))

	// todo cleanup
	msg, err := message.NewMessage(
		id,
		unsigned.Previous(),
		unsigned.Sequence(),
		unsigned.Author(),
		unsigned.Feed(),
		unsigned.Timestamp(),
		unsigned.Content(),
		message.NewRawMessage(raw),
	)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "could not create a message")
	}

	return msg, nil
}
