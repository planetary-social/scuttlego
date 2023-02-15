package commands

import (
	"io"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type BlobCreator interface {
	Create(r io.Reader) (refs.Blob, error)
}

type CreateBlob struct {
	reader io.Reader
}

func NewCreateBlob(reader io.Reader) (CreateBlob, error) {
	if reader == nil {
		return CreateBlob{}, errors.New("zero value of reader")
	}
	return CreateBlob{reader: reader}, nil
}

func (c CreateBlob) Reader() io.Reader {
	return c.reader
}

func (c CreateBlob) IsZero() bool {
	return c == CreateBlob{}
}

type CreateBlobHandler struct {
	creator BlobCreator
}

func NewCreateBlobHandler(
	creator BlobCreator,
) *CreateBlobHandler {
	return &CreateBlobHandler{
		creator: creator,
	}
}

func (h *CreateBlobHandler) Handle(cmd CreateBlob) (refs.Blob, error) {
	if cmd.IsZero() {
		return refs.Blob{}, errors.New("zero value of cmd")
	}
	return h.creator.Create(cmd.Reader())
}
