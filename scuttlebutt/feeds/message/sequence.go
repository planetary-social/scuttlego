package message

import "github.com/boreq/errors"

var FirstSequence = Sequence{1}

type Sequence struct {
	s int
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
	return s == FirstSequence
}

func (s Sequence) ComesDirectlyBefore(o Sequence) bool {
	return s.s == o.s-1
}

func (s Sequence) Int() int {
	return s.s
}

func (s Sequence) ComesAfter(o Sequence) bool {
	return s.s > o.s
}

func (s Sequence) Next() Sequence {
	return Sequence{s.s + 1}
}
