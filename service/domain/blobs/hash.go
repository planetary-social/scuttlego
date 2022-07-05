package blobs

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"hash"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type Hasher struct {
	h hash.Hash
}

func NewHasher() Hasher {
	return Hasher{
		h: sha256.New(),
	}
}

func (h Hasher) Write(p []byte) (n int, err error) {
	return h.h.Write(p)
}

func (h Hasher) Sum(b []byte) []byte {
	return h.h.Sum(b)
}

func (h Hasher) Reset() {
	h.h.Reset()
}

func (h Hasher) Size() int {
	return h.h.Size()
}

func (h Hasher) BlockSize() int {
	return h.h.BlockSize()
}

func (h Hasher) SumRef() (refs.Blob, error) {
	computedHash := h.h.Sum(nil)
	return refs.NewBlob("&" + base64.StdEncoding.EncodeToString(computedHash) + ".sha256")
}

func Verify(id refs.Blob, h hash.Hash) error {
	if !bytes.Equal(id.Bytes(), h.Sum(nil)) {
		return errors.New("invalid blob hash")
	}
	return nil
}
