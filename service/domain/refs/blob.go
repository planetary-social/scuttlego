package refs

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/boreq/errors"
)

const (
	blobPrefix = "&"
	blobSuffix = ".sha256"

	blobHashLength = 32
)

type Blob struct {
	s string
	b []byte
}

func NewBlob(s string) (Blob, error) {
	if !strings.HasPrefix(s, blobPrefix) {
		return Blob{}, errors.New("invalid prefix")
	}

	if !strings.HasSuffix(s, blobSuffix) {
		return Blob{}, errors.New("invalid suffix")
	}

	noSuffixAndPrefix := s[len(blobPrefix) : len(s)-len(blobSuffix)]

	b, err := base64.StdEncoding.DecodeString(noSuffixAndPrefix)
	if err != nil {
		return Blob{}, errors.Wrap(err, "invalid base64")
	}

	if l := len(b); l != blobHashLength {
		return Blob{}, fmt.Errorf("invalid hash length '%d'", l)
	}

	return Blob{s, b}, nil
}

func MustNewBlob(s string) Blob {
	r, err := NewBlob(s)
	if err != nil {
		panic(err)
	}
	return r
}

func (m Blob) Bytes() []byte {
	return m.b
}

func (m Blob) String() string {
	return m.s
}

func (m Blob) IsZero() bool {
	return len(m.b) == 0
}

func (m Blob) Equal(o Message) bool {
	return bytes.Equal(m.b, o.b)
}
