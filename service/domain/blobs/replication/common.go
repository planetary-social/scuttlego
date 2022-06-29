package replication

import (
	"context"
	"io"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/blobs"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/transport"
)

type WantListStorage interface {
	GetWantList() (blobs.WantList, error)
}

type Downloader interface {
	OnHasReceived(ctx context.Context, peer transport.Peer, blob refs.Blob, size blobs.Size)
}

var ErrBlobNotFound = errors.New("blob not found")

type BlobStorage interface {
	Store(id refs.Blob, r io.Reader) error

	// Size returns the size of the blob. If the blob is not found it returns
	// ErrBlobNotFound.
	Size(id refs.Blob) (blobs.Size, error)
}
