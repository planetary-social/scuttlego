package blobs

import "github.com/boreq/errors"

const wantDistanceLocal = 1

// WantDistance represents how far away the peer who wants a blob is.
//
// The 1 means the responder wants this blob themselves. 2 means they know
// someone who wants the blob. 3 means they know someone who knows someone who
// wants the blob, and so on.
type WantDistance struct {
	distance int
}

func NewWantDistance(distance int) (WantDistance, error) {
	if distance <= 0 {
		return WantDistance{}, errors.New("distance must be positive")
	}

	return WantDistance{distance: distance}, nil
}

func NewWantDistanceLocal() WantDistance {
	return WantDistance{distance: wantDistanceLocal}
}

func MustNewWantDistance(distance int) WantDistance {
	v, err := NewWantDistance(distance)
	if err != nil {
		panic(err)
	}
	return v
}

// Int returns a positive number.
func (s WantDistance) Int() int {
	return s.distance
}

func (s WantDistance) IsZero() bool {
	return s == WantDistance{}
}
