package queries

import (
	"io"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type BlobStorage interface {
	Get(id refs.Blob) (io.ReadCloser, error)
	Size(id refs.Blob) (blobs.Size, error)
}

type GetBlob struct {
	id refs.Blob

	// Size is the expected size of the blob in bytes. If the blob is not
	// exactly this size then an error is returned.
	size *blobs.Size

	// Max is the Maximum size of the blob in bytes. If the blob is larger then
	// an error is returned.
	max *blobs.Size
}

func NewGetBlob(id refs.Blob, size *blobs.Size, max *blobs.Size) (GetBlob, error) {
	if id.IsZero() {
		return GetBlob{}, errors.New("zero value of id")
	}
	return GetBlob{id: id, size: size, max: max}, nil
}

func (g GetBlob) Id() refs.Blob {
	return g.id
}

func (g GetBlob) Size() *blobs.Size {
	return g.size
}

func (g GetBlob) Max() *blobs.Size {
	return g.max
}

func (g GetBlob) IsZero() bool {
	return g.id.IsZero()
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
	if query.IsZero() {
		return nil, errors.New("zero value of query")
	}

	if query.Size() != nil || query.Max() != nil {
		blobSize, err := h.storage.Size(query.Id())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get the blob size")
		}

		if query.Size() != nil {
			if blobSize != *query.Size() {
				return nil, errors.New("blob size doesn't match the provided size")
			}
		}

		if query.Max() != nil {
			if blobSize.Above(*query.Max()) {
				return nil, errors.New("blob is larger than the provided max size")
			}
		}
	}

	return h.storage.Get(query.Id())
}
