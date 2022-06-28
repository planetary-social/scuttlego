package queries

import (
	"io"

	"github.com/planetary-social/go-ssb/service/domain/refs"
)

type BlobStorage interface {
	GetBlob(id refs.Blob) (io.ReadCloser, error)
}

type GetBlob struct {
	Id refs.Blob
}

type GetBlobHandler struct {
	storage BlobStorage
}

func NewGetBlobHandler(storage BlobStorage) (*GetBlobHandler, error) {
	return &GetBlobHandler{
		storage: storage,
	}, nil
}

func (h *GetBlobHandler) Handle(query GetBlob) (io.ReadCloser, error) {
	return h.storage.GetBlob(query.Id)
}
