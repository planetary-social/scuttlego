package formats

import (
	"fmt"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/go-ssb/service/domain/feeds"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
)

type Marshaler interface {
	Marshal(content message.MessageContent) ([]byte, error)
	Unmarshal(b []byte) (message.MessageContent, error)
}

type RawMessageIdentifier struct {
	formats []feeds.FeedFormat
}

func NewRawMessageIdentifier(formats []feeds.FeedFormat) *RawMessageIdentifier {
	return &RawMessageIdentifier{
		formats: formats,
	}
}

func (i RawMessageIdentifier) IdentifyRawMessage(raw message.RawMessage) (message.Message, error) {
	var result error

	for _, format := range i.formats {
		msg, err := format.Verify(raw)
		if err == nil {
			return msg, nil
		}

		result = multierror.Append(result, err)
	}

	return message.Message{}, errors.Wrap(result, "unknown message")
}

// messageHMACLength is implied to be constant due to an assumption that this key is used as an HMAC key when calling
// libsodium's functions.
const messageHMACLength = 32

// MessageHMAC is mainly used for test networks. It is applied to messages to make them incompatible with the main SSB
// network.
// https://github.com/ssb-js/ssb-validate#state--validateappendstate-hmac_key-msg
type MessageHMAC struct {
	b []byte
}

func NewMessageHMAC(b []byte) (MessageHMAC, error) {
	if len(b) == 0 {
		return NewDefaultMessageHMAC(), nil
	}

	if len(b) != messageHMACLength {
		return MessageHMAC{}, fmt.Errorf("invalid message HMAC length, must be '%d'", messageHMACLength)
	}

	buf := make([]byte, messageHMACLength)
	copy(buf, b)
	return MessageHMAC{buf}, nil
}

func NewDefaultMessageHMAC() MessageHMAC {
	return MessageHMAC{nil}
}

func (k MessageHMAC) Bytes() []byte {
	if k.IsZero() {
		return nil
	}

	tmp := make([]byte, len(k.b))
	copy(tmp, k.b)
	return tmp
}

func (k MessageHMAC) IsZero() bool {
	return len(k.b) == 0
}
