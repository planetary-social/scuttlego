package content

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

var ErrUnknownContent = errors.New("unknown content")

type Marshaler interface {
	Marshal(content known.KnownMessageContent) (message.RawMessageContent, error)

	// Unmarshal returns ErrUnknownContent if the content isn't known.
	Unmarshal(b message.RawMessageContent) (known.KnownMessageContent, error)
}

type BlobScanner interface {
	Scan(rawContent message.RawMessageContent) ([]refs.Blob, error)
}

type Parser struct {
	marshaler   Marshaler
	blobScanner BlobScanner
}

func NewParser(marshaler Marshaler, blobScanner BlobScanner) *Parser {
	return &Parser{marshaler: marshaler, blobScanner: blobScanner}
}

func (p *Parser) Parse(raw message.RawMessageContent) (message.Content, error) {
	knownContent, err := p.getKnownContent(raw)
	if err != nil {
		return message.Content{}, errors.Wrap(err, "could not get known content")
	}

	referencedBlobs, err := p.blobScanner.Scan(raw)
	if err != nil {
		return message.Content{}, errors.Wrap(err, "error scanning for blobs")
	}

	return message.NewContent(raw, knownContent, referencedBlobs)
}

func (p *Parser) getKnownContent(rawContent message.RawMessageContent) (known.KnownMessageContent, error) {
	knownContent, err := p.marshaler.Unmarshal(rawContent)
	if err != nil {
		if errors.Is(err, ErrUnknownContent) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "could not unmarshal message content")
	}
	return knownContent, nil
}
