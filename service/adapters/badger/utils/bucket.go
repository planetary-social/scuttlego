package utils

import (
	"bytes"

	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
)

type Bucket struct {
	tx     *badger.Txn
	prefix Key
}

func NewBucket(tx *badger.Txn, prefix Key) (Bucket, error) {
	if tx == nil {
		return Bucket{}, errors.New("nil tx")
	}

	if prefix.IsZero() {
		return Bucket{}, errors.New("zero value of prefix")
	}

	return Bucket{tx: tx, prefix: prefix}, nil
}

func MustNewBucket(tx *badger.Txn, prefix Key) Bucket {
	v, err := NewBucket(tx, prefix)
	if err != nil {
		panic(err)
	}
	return v
}

func (b Bucket) Set(key, val []byte) error {
	targetKey, err := b.targetKey(key)
	if err != nil {
		return errors.Wrap(err, "error creating a target key")
	}

	return b.tx.Set(targetKey, val)
}

func (b Bucket) Delete(key []byte) error {
	targetKey, err := b.targetKey(key)
	if err != nil {
		return errors.Wrap(err, "error creating a target key")
	}

	return b.tx.Delete(targetKey)
}

func (b Bucket) Get(key []byte) (*badger.Item, error) {
	targetKey, err := b.targetKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "error creating a target key")
	}

	return b.tx.Get(targetKey)
}

func (b Bucket) ForEach(fn func(item *badger.Item) error) error {
	it := b.tx.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	prefix := b.prefix.Bytes()
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		if err := fn(it.Item()); err != nil {
			return errors.Wrap(err, "function returned an error")
		}
	}
	return nil
}

func (b Bucket) Iterator() *BucketIterator {
	return NewBucketIterator(b)
}

func (b Bucket) IteratorWithModifiedOptions(fn func(options *badger.IteratorOptions)) *BucketIterator {
	return NewBucketIteratorWithModifiedOptions(b, fn)
}

func (b Bucket) ChildBucket(component KeyComponent) Bucket {
	return MustNewBucket(b.tx, b.prefix.Append(component))
}

func (b Bucket) DeleteBucket() error {
	if err := b.ForEach(func(item *badger.Item) error { // todo don't prefech values? // todo do it faster somehow?
		if err := b.tx.Delete(item.Key()); err != nil {
			return errors.Wrap(err, "delete error")
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "foreach error")
	}

	return nil
}

func (b Bucket) targetKey(key []byte) ([]byte, error) {
	keyComponent, err := NewKeyComponent(key)
	if err != nil {
		return nil, errors.Wrap(err, "error creating a key component")
	}
	return b.prefix.Append(keyComponent).Bytes(), nil
}

func (b Bucket) KeyInBucket(item *badger.Item) (KeyComponent, error) {
	itemKey, err := NewKeyFromBytes(item.KeyCopy(nil)) // todo copy only sometimes or later?
	if err != nil {
		return KeyComponent{}, errors.Wrap(err, "error parsing the key")
	}

	if len(itemKey.components) != len(b.prefix.components)+1 {
		return KeyComponent{}, errors.New("invalid item key length")
	}

	for i := range b.prefix.components {
		if !bytes.Equal(itemKey.components[i].b, b.prefix.components[i].b) {
			return KeyComponent{}, errors.New("invalid item key component")
		}
	}

	return itemKey.components[len(itemKey.components)-1], nil
}

func (b Bucket) IsEmpty() bool {
	it := b.Iterator()
	defer it.Close()

	it.Rewind()
	return !it.Valid()
}
