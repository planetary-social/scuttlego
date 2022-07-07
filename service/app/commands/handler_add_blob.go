package commands

import (
	"io"

	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type BlobCreator interface {
	Create(r io.Reader) (refs.Blob, error)
}

type CreateBlob struct {
	Reader io.Reader
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
	return h.creator.Create(cmd.Reader)
}
