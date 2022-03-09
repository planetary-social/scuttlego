package transport

import (
	"encoding/json"
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/content"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
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

func (m *Marshaler) Marshal(content message.MessageContent) ([]byte, error) {
	typ := content.Type()

	mapping, ok := m.mappings[typ]
	if !ok {
		return nil, fmt.Errorf("no mapping for '%s'", typ)
	}

	return mapping.Marshal(content)
}

func (m *Marshaler) Unmarshal(b []byte) (message.MessageContent, error) {
	logger := m.logger.WithField("content", string(b))

	typ, err := m.identifyContentType(b)
	if err != nil {
		logger.WithError(err).Error("failed to identify message content type")
		return content.NewUnknown(b)
	}

	mapping, ok := m.mappings[typ]
	if !ok {
		return content.NewUnknown(b)
	}

	cnt, err := mapping.Unmarshal(b)
	if err != nil {
		logger.WithField("typ", typ).WithError(err).Error("mapping returned an error")
		return nil, errors.Wrapf(err, "mapping '%s' returned an error", typ)
	}

	return cnt, nil
}

func (m *Marshaler) identifyContentType(b []byte) (message.MessageContentType, error) {
	var typ messageContentType
	if err := json.Unmarshal(b, &typ); err != nil {
		return "", errors.Wrap(err, "json unmarshal of message content type failed")
	}
	if typ.MessageContentType == string((content.Unknown{}).Type()) {
		return "", errors.New("empty content type")
	}
	return message.MessageContentType(typ.MessageContentType), nil
}

type messageContentType struct {
	MessageContentType string `json:"type"`
}
