package message

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type Content struct {
	raw             RawMessageContent
	known           known.KnownMessageContent
	referencedBlobs []refs.Blob
}

func NewContent(
	raw RawMessageContent,
	known known.KnownMessageContent,
	referencedBlobs []refs.Blob,
) (Content, error) {
	if raw.IsZero() {
		return Content{}, errors.New("zero value of raw")
	}
	return Content{
		raw:             raw,
		known:           known,
		referencedBlobs: referencedBlobs,
	}, nil
}

func MustNewContent(
	raw RawMessageContent,
	known known.KnownMessageContent,
	referencedBlobs []refs.Blob,
) Content {
	v, err := NewContent(raw, known, referencedBlobs)
	if err != nil {
		panic(err)
	}
	return v
}

func (c Content) Raw() RawMessageContent {
	return c.raw
}

func (c Content) KnownContent() (known.KnownMessageContent, bool) {
	if c.known == nil {
		return nil, false
	}
	return c.known, true
}

func (c Content) ReferencedBlobs() []refs.Blob {
	return c.referencedBlobs
}

func (c Content) IsZero() bool {
	return c.raw.IsZero()
}
