package blobs

import (
	"bytes"
	"crypto/sha256"
	"hash"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

type Verifier struct {
	id   refs.Blob
	hash hash.Hash
}

func NewVerifier(id refs.Blob) *Verifier {
	return &Verifier{
		id:   id,
		hash: sha256.New(),
	}
}

func (h *Verifier) Write(p []byte) (n int, err error) {
	return h.hash.Write(p)
}

func (h *Verifier) Verify() error {
	computedHash := h.hash.Sum(nil)
	if !bytes.Equal(computedHash, h.id.Bytes()) {
		return errors.New("invalid blob hash")
	}
	return nil
}
