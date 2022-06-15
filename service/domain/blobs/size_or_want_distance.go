package blobs

import "github.com/boreq/errors"

type SizeOrWantDistance struct {
	size     *Size
	distance *WantDistance
}

func NewSizeOrWantDistance(v int64) (SizeOrWantDistance, error) {
	if v == 0 {
		return SizeOrWantDistance{}, errors.New("0 is neither a distance nor the size")
	}

	if v > 0 {
		size, err := NewSize(v)
		if err != nil {
			return SizeOrWantDistance{}, errors.Wrap(err, "could not create a size")
		}

		return SizeOrWantDistance{size: &size}, nil
	}

	if v < 0 {
		distance, err := NewWantDistance(-int(v))
		if err != nil {
			return SizeOrWantDistance{}, errors.Wrap(err, "could not create a distance")
		}

		return SizeOrWantDistance{distance: &distance}, nil
	}

	panic("logic error") // as the famous last words go: this should not be possible
}

func NewSizeOrWantDistanceContainingWantDistance(distance WantDistance) (SizeOrWantDistance, error) {
	return NewSizeOrWantDistance(-int64(distance.Int()))
}

func (s SizeOrWantDistance) Size() (Size, bool) {
	if s.size != nil {
		return *s.size, true
	}
	return Size{}, false
}

func (s SizeOrWantDistance) WantDistance() (WantDistance, bool) {
	if s.distance != nil {
		return *s.distance, true
	}
	return WantDistance{}, false
}

func (s SizeOrWantDistance) IsZero() bool {
	return s == SizeOrWantDistance{}
}
