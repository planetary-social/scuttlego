package utils

import (
	"encoding/binary"

	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
)

type Counter struct {
	bucket Bucket
	key    KeyComponent
}

func NewCounter(bucket Bucket, key KeyComponent) (Counter, error) {
	if key.IsZero() {
		return Counter{}, errors.New("zero value of key")
	}

	return Counter{
		bucket: bucket,
		key:    key,
	}, nil
}

func (c Counter) Increment() error {
	existingSequence, err := c.get()
	if err != nil {
		return errors.Wrap(err, "error getting existing sequence")
	}

	return c.set(existingSequence + 1)
}

func (c Counter) Decrement() error {
	existingSequence, err := c.get()
	if err != nil {
		return errors.Wrap(err, "error getting existing sequence")
	}

	return c.set(existingSequence - 1)
}

func (c Counter) Get() (uint64, error) {
	return c.get()
}

func (c Counter) get() (uint64, error) {
	item, err := c.bucket.Get(c.key.Bytes())
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return 0, nil
		}
		return 0, errors.Wrap(err, "error getting sequence")
	}

	var existingSequence uint64

	if err := item.Value(func(val []byte) error {
		tmp, err := c.unmarshal(val)
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

func (c Counter) set(v uint64) error {
	if err := c.bucket.Set(c.key.Bytes(), c.marshal(v)); err != nil {
		return errors.Wrap(err, "set error")
	}
	return nil
}

func (c Counter) marshal(v uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return b
}

func (c Counter) unmarshal(b []byte) (uint64, error) {
	if len(b) != 8 {
		return 0, errors.New("invalid length")
	}
	return binary.LittleEndian.Uint64(b), nil
}
