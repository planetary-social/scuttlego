package blobs

import "github.com/planetary-social/go-ssb/service/domain/refs"

// BlobReferencer is an interface implemented by some message contents e.g.
// about or post. Messages implementing this interface reference blobs.
type BlobReferencer interface {
	Blobs() []refs.Blob
}
