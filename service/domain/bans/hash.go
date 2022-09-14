package bans

import (
	"crypto/sha256"
	"fmt"
)

type Hash struct {
	h []byte
}

func NewHash(h []byte) (Hash, error) {
	if l := len(h); l != sha256.Size {
		return Hash{}, fmt.Errorf("invalid hash length '%d'", l)
	}
	return Hash{h: h}, nil
}

func MustNewHash(h []byte) Hash {
	v, err := NewHash(h)
	if err != nil {
		panic(err)
	}
	return v
}

func (h Hash) Bytes() []byte {
	tmp := make([]byte, len(h.h))
	copy(tmp, h.h)
	return tmp
}

func (h Hash) IsZero() bool {
	return len(h.h) == 0
}
