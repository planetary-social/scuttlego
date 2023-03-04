package transport

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
)

type Marshaler struct {
	mappings MessageContentMappings
	logger   logging.Logger
}

func NewMarshaler(mappings MessageContentMappings, logger logging.Logger) (*Marshaler, error) {
	for typ := range mappings {
		if typ.IsZero() {
			return nil, errors.New("zero value of message content type")
		}
	}

	return &Marshaler{
		mappings: mappings,
		logger:   logger,
	}, nil
}

func (m *Marshaler) Marshal(content known.KnownMessageContent) (message.RawContent, error) {
	typ := content.Type()

	mapping, ok := m.mappings[typ]
	if !ok {
		return message.RawContent{}, fmt.Errorf("no mapping for '%s'", typ)
	}

	b, err := mapping.Marshal(content)
	if err != nil {
		return message.RawContent{}, errors.Wrap(err, "invalid content")
	}

	return message.NewRawContent(b)
}

func (m *Marshaler) Unmarshal(b message.RawContent) (known.KnownMessageContent, error) {
	logger := m.logger.WithField("content", string(b.Bytes()))

	typ, err := m.identifyContentType(b)
	if err != nil {
		// todo how to deal with box
		if !strings.HasSuffix(string(b.Bytes()), ".box\"") {
			logger.Error().WithError(err).Message("failed to identify message content type")
		}
		return nil, content.ErrUnknownContent
	}

	mapping, ok := m.mappings[typ]
	if !ok {
		return nil, content.ErrUnknownContent
	}

	knownContent, err := mapping.Unmarshal(b.Bytes())
	if err != nil {
		logger.Error().WithField("typ", typ).WithError(err).Message("mapping returned an error")
		return nil, content.ErrUnknownContent
	}

	return knownContent, nil
}

func (m *Marshaler) identifyContentType(b message.RawContent) (known.MessageContentType, error) {
	var typ MessageContentType
	if err := json.Unmarshal(b.Bytes(), &typ); err != nil {
		return "", errors.Wrap(err, "json unmarshal of message content type failed")
	}
	if typ.MessageContentType == "" {
		return "", errors.New("empty content type")
	}
	return known.MessageContentType(typ.MessageContentType), nil
}

type MessageContentType struct {
	MessageContentType string `json:"type"`
}

func NewMessageContentType(messageContent known.KnownMessageContent) MessageContentType {
	return MessageContentType{string(messageContent.Type())}
}
