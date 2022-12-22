package utils

import (
	"encoding/binary"

	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
)

type Sequence struct {
	bucket Bucket
	key    KeyComponent
}

func NewSequence(bucket Bucket, key KeyComponent) (Sequence, error) {
	if key.IsZero() {
		return Sequence{}, errors.New("zero value of key")
	}

	return Sequence{
		bucket: bucket,
		key:    key,
	}, nil
}

// Next returns the next sequence and saves it. First call to next returns 1.
func (s Sequence) Next() (uint64, error) {
	existingSequence, err := s.get()
	if err != nil {
		return 0, errors.Wrap(err, "error getting existing sequence")
	}

	newSequence := existingSequence + 1

	if err := s.set(newSequence); err != nil {
		return 0, errors.Wrap(err, "error setting new sequence")
	}

	return newSequence, nil
}

func (s Sequence) Get() (uint64, error) {
	return s.get()
}

func (s Sequence) Set(v uint64) error {
	return s.set(v)
}

func (s Sequence) get() (uint64, error) {
	item, err := s.bucket.Get(s.key.Bytes())
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return 0, nil
		}
		return 0, errors.Wrap(err, "error getting sequence")
	}

	var existingSequence uint64

	if err := item.Value(func(val []byte) error {
		tmp, err := s.unmarshal(val)
		if err != nil {
			return errors.Wrap(err, "unmarshal error")
		}
		existingSequence = tmp
		return err
	}); err != nil {
		return 0, errors.Wrap(err, "error getting value")
	}

	return existingSequence, nil
}

func (s Sequence) set(v uint64) error {
	if err := s.bucket.Set(s.key.Bytes(), s.marshal(v)); err != nil {
		return errors.Wrap(err, "set error")
	}
	return nil
}

func (s Sequence) marshal(v uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return b
}

func (s Sequence) unmarshal(b []byte) (uint64, error) {
	if len(b) != 8 {
		return 0, errors.New("invalid length")
	}
	return binary.LittleEndian.Uint64(b), nil
}
