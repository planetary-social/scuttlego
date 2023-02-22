package formats

import (
	"encoding/json"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	ssbrefs "github.com/ssbc/go-ssb-refs"
	"github.com/ssbc/go-ssb/message/legacy"
)

type Scuttlebutt struct {
	parser ContentParser
	hmac   MessageHMAC
}

func NewScuttlebutt(
	parser ContentParser,
	hmac MessageHMAC,
) *Scuttlebutt {
	return &Scuttlebutt{
		parser: parser,
		hmac:   hmac,
	}
}

func (s *Scuttlebutt) Verify(raw message.RawMessage) (message.Message, error) {
	ssbRef, ssbMessage, err := legacy.Verify(raw.Bytes(), s.convertHMAC())
	if err != nil {
		return message.Message{}, errors.Wrap(err, "verification failed")
	}

	msgWithoutId, err := s.convert(ssbMessage, raw)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "error converting message")
	}

	id, err := refs.NewMessage(ssbRef.Sigil())
	if err != nil {
		return message.Message{}, errors.Wrap(err, "invalid message id")
	}

	msg, err := message.NewMessageFromMessageWithoutId(id, msgWithoutId)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "could not create a message")
	}

	return msg, nil
}

func (s *Scuttlebutt) Sign(unsigned message.UnsignedMessage, private identity.Private) (message.Message, error) {
	var previous *ssbrefs.MessageRef
	if unsigned.Previous() != nil {
		tmp, err := ssbrefs.NewMessageRefFromBytes(unsigned.Previous().Bytes(), ssbrefs.RefAlgoMessageSSB1)
		if err != nil {
			return message.Message{}, errors.New("could not create a ref")
		}
		previous = &tmp
	}

	msgToSign := legacy.LegacyMessage{
		Previous:  previous,
		Author:    unsigned.Author().String(),
		Sequence:  int64(unsigned.Sequence().Int()),
		Timestamp: unsigned.Timestamp().UnixMilli(),
		Hash:      "sha256",
		Content:   json.RawMessage(unsigned.Content().Bytes()),
	}

	_, raw, err := msgToSign.Sign(private.PrivateKey(), s.convertHMAC())
	if err != nil {
		return message.Message{}, errors.Wrap(err, "could not sign a message")
	}

	rawMessage, err := message.NewRawMessage(raw)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "could not create a raw message")
	}

	return s.Verify(rawMessage)
}

func (s *Scuttlebutt) Load(verifiedRawMessage message.VerifiedRawMessage) (message.MessageWithoutId, error) {
	raw, err := message.NewRawMessage(verifiedRawMessage.Bytes())
	if err != nil {
		return message.MessageWithoutId{}, errors.Wrap(err, "could not create raw message")
	}

	var dmsg legacy.DeserializedMessage
	if err := json.Unmarshal(verifiedRawMessage.Bytes(), &dmsg); err != nil {
		return message.MessageWithoutId{}, errors.Wrap(err, "json unmarshal failed")
	}

	return s.convert(dmsg, raw)
}

func (s *Scuttlebutt) Peek(raw message.RawMessage) (feeds.PeekedMessage, error) {
	var dmsg peekedSsbMessage
	if err := json.Unmarshal(raw.Bytes(), &dmsg); err != nil {
		return feeds.PeekedMessage{}, errors.Wrap(err, "json unmarshal failed")
	}

	feed, err := refs.NewFeed(dmsg.Author)
	if err != nil {
		return feeds.PeekedMessage{}, errors.New("error creating a feed ref")
	}

	sequence, err := message.NewSequence(dmsg.Sequence)
	if err != nil {
		return feeds.PeekedMessage{}, errors.New("error creating a sequence")
	}

	return feeds.NewPeekedMessage(feed, sequence, raw)
}

type peekedSsbMessage struct {
	Author   string `json:"author"`
	Sequence int    `json:"sequence"`
}

func (s *Scuttlebutt) convert(ssbMessage legacy.DeserializedMessage, raw message.RawMessage) (message.MessageWithoutId, error) {
	var previous *refs.Message
	if ssbMessage.Previous != nil {
		tmp, err := refs.NewMessage(ssbMessage.Previous.Sigil())
		if err != nil {
			return message.MessageWithoutId{}, errors.Wrap(err, "invalid previous message id")
		}
		previous = &tmp
	}

	sequence, err := message.NewSequence(int(ssbMessage.Sequence))
	if err != nil {
		return message.MessageWithoutId{}, errors.Wrap(err, "invalid sequence")
	}

	author, err := refs.NewIdentity(ssbMessage.Author.Sigil())
	if err != nil {
		return message.MessageWithoutId{}, errors.Wrap(err, "invalid author")
	}

	feed, err := refs.NewFeed(ssbMessage.Author.Sigil())
	if err != nil {
		return message.MessageWithoutId{}, errors.Wrap(err, "invalid feed")
	}

	timestamp := time.UnixMilli(int64(ssbMessage.Timestamp))

	rawMessageContent, err := message.NewRawContent(ssbMessage.Content)
	if err != nil {
		return message.MessageWithoutId{}, errors.Wrap(err, "could not create raw message content")
	}

	content, err := s.parser.Parse(rawMessageContent)
	if err != nil {
		return message.MessageWithoutId{}, errors.Wrap(err, "could not parse the content")
	}

	msg, err := message.NewMessageWithoutId(
		previous,
		sequence,
		author,
		feed,
		timestamp,
		content,
		raw,
	)
	if err != nil {
		return message.MessageWithoutId{}, errors.Wrap(err, "could not create a message")
	}

	return msg, nil
}

func (s *Scuttlebutt) convertHMAC() *[32]byte {
	if s.hmac.IsZero() {
		return nil
	}

	return (*[32]byte)(s.hmac.Bytes())
}
