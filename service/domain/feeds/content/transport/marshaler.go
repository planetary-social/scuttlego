package transport

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/feeds/content"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
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

func (m *Marshaler) Marshal(content content.KnownMessageContent) (message.RawMessageContent, error) {
	typ := content.Type()

	mapping, ok := m.mappings[typ]
	if !ok {
		return message.RawMessageContent{}, fmt.Errorf("no mapping for '%s'", typ)
	}

	b, err := mapping.Marshal(content)
	if err != nil {
		return message.RawMessageContent{}, errors.Wrap(err, "invalid content")
	}

	return message.NewRawMessageContent(b)
}

func (m *Marshaler) Unmarshal(b message.RawMessageContent) (content.KnownMessageContent, error) {
	logger := m.logger.WithField("content", string(b.Bytes()))

	typ, err := m.identifyContentType(b)
	if err != nil {
		// todo how to deal with box
		if !strings.HasSuffix(string(b.Bytes()), ".box\"") {
			logger.WithError(err).Error("failed to identify message content type")
		}
		return content.NewUnknown(b.Bytes())
	}

	mapping, ok := m.mappings[typ]
	if !ok {
		return content.NewUnknown(b.Bytes())
	}

	cnt, err := mapping.Unmarshal(b.Bytes())
	if err != nil {
		logger.WithField("typ", typ).WithError(err).Error("mapping returned an error")
		return nil, errors.Wrapf(err, "mapping '%s' returned an error", typ)
	}

	return cnt, nil
}

func (m *Marshaler) identifyContentType(b message.RawMessageContent) (content.MessageContentType, error) {
	var typ messageContentType
	if err := json.Unmarshal(b.Bytes(), &typ); err != nil {
		return "", errors.Wrap(err, "json unmarshal of message content type failed")
	}
	if typ.MessageContentType == "" {
		return "", errors.New("empty content type")
	}
	return content.MessageContentType(typ.MessageContentType), nil
}

type messageContentType struct {
	MessageContentType string `json:"type"`
}
