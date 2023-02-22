package formats

import (
	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
)

type ContentParser interface {
	Parse(raw message.RawContent) (message.Content, error)
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

func (i RawMessageIdentifier) PeekRawMessage(raw message.RawMessage) (feeds.PeekedMessage, error) {
	var result error

	for _, format := range i.formats {
		msg, err := format.Peek(raw)
		if err == nil {
			return msg, nil
		}

		result = multierror.Append(result, err)
	}

	return feeds.PeekedMessage{}, errors.Wrap(result, "unknown message")
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
