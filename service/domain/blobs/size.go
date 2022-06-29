package blobs

import "github.com/boreq/errors"

const maxBlobSize = 5 * 1024 * 1024

type Size struct {
	sizeInBytes int64
}

func NewSize(sizeInBytes int64) (Size, error) {
	if sizeInBytes <= 0 {
		return Size{}, errors.New("size must be positive")
	}

	return Size{sizeInBytes: sizeInBytes}, nil
}

func MustNewSize(sizeInBytes int64) Size {
	v, err := NewSize(sizeInBytes)
	if err != nil {
		panic(err)
	}
	return v
}

func MaxBlobSize() Size {
	return MustNewSize(maxBlobSize)
}

func (s Size) InBytes() int64 {
	return s.sizeInBytes
}

func (s Size) Above(other Size) bool {
	return s.sizeInBytes > other.sizeInBytes
}

func (s Size) IsZero() bool {
	return s == Size{}
}
