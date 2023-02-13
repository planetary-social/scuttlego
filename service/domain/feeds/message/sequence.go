package message

import "github.com/boreq/errors"

const firstSequence = 1

type Sequence struct {
	s int
}

func NewFirstSequence() Sequence {
	return Sequence{firstSequence}
}

func NewSequence(s int) (Sequence, error) {
	if s <= 0 {
		return Sequence{}, errors.New("sequence must be positive")
	}
	return Sequence{s}, nil
}

func MustNewSequence(s int) Sequence {
	r, err := NewSequence(s)
	if err != nil {
		panic(err)
	}
	return r
}

func (s Sequence) IsFirst() bool {
	return s.s == firstSequence
}

func (s Sequence) ComesDirectlyBefore(o Sequence) bool {
	return s.s == o.s-1
}

func (s Sequence) ComesAfter(o Sequence) bool {
	return s.s > o.s
}

func (s Sequence) Next() Sequence {
	return MustNewSequence(s.s + 1)
}

func (s Sequence) Previous() (Sequence, bool) {
	prevSequence := s.s - 1
	if firstSequence > prevSequence {
		return Sequence{}, false
	}
	return MustNewSequence(prevSequence), true
}

func (s Sequence) Int() int {
	return s.s
}

func (s Sequence) IsZero() bool {
	return s == Sequence{}
}
