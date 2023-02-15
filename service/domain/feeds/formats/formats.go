package formats

import (
	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
)

type Marshaler interface {
	Marshal(content content.KnownMessageContent) (message.RawMessageContent, error)
	Unmarshal(b message.RawMessageContent) (content.KnownMessageContent, error)
}

type RawMessageIdentifier struct {
	formats []feeds.FeedFormat
}

func NewRawMessageIdentifier(formats []feeds.FeedFormat) *RawMessageIdentifier {
	return &RawMessageIdentifier{
		formats: formats,
	}
}

func (i RawMessageIdentifier) VerifyRawMessage(raw message.RawMessage) (message.Message, error) {
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

func (i RawMessageIdentifier) LoadRawMessage(raw message.VerifiedRawMessage) (message.MessageWithoutId, error) {
	var result error

	for _, format := range i.formats {
		msg, err := format.Load(raw)
		if err == nil {
			return msg, nil
		}

		result = multierror.Append(result, err)
	}

	return message.MessageWithoutId{}, errors.Wrap(result, "unknown message")
}
