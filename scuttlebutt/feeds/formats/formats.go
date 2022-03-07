package formats

import (
	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
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
