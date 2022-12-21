package utils

import (
	"bytes"
	"fmt"
	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"io"
	"math"
)

const maxKeyComponentLength = math.MaxUint8

type KeyComponent struct {
	b []byte
}

func NewKeyComponent(b []byte) (KeyComponent, error) {
	if len(b) == 0 {
		return KeyComponent{}, errors.New("empty key component")
	}

	if l := len(b); l > math.MaxUint8 {
		return KeyComponent{}, fmt.Errorf("key component too long: %d", l)
	}

	return KeyComponent{b: b}, nil
}

func MustNewKeyComponent(b []byte) KeyComponent {
	v, err := NewKeyComponent(b)
	if err != nil {
		panic(err)
	}
	return v
}

func (k KeyComponent) Bytes() []byte {
	return k.b
}

func (k KeyComponent) IsZero() bool {
	return len(k.b) == 0
}

type Key struct {
	components []KeyComponent
}

func NewKey(components ...KeyComponent) (Key, error) {
	if len(components) == 0 {
		return Key{}, errors.New("no key components given")
	}

	for _, component := range components {
		if component.IsZero() {
			return Key{}, errors.New("zero value of key component")
		}
	}

	return Key{components: components}, nil
}

func MustNewKey(components ...KeyComponent) Key {
	v, err := NewKey(components...)
	if err != nil {
		panic(err)
	}
	return v
}

func NewKeyFromBytes(b []byte) (Key, error) {
	buf := bytes.NewBuffer(b)
	var components []KeyComponent

	for {
		nextSequenceLen, err := buf.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return Key{}, errors.Wrap(err, "error reading the next sequence length")
		}

		nextSequenceBuf := make([]byte, nextSequenceLen)
		n, err := buf.Read(nextSequenceBuf)
		if err != nil {
			return Key{}, errors.Wrap(err, "error reading the next sequence")
		}

		if n != int(nextSequenceLen) {
			return Key{}, fmt.Errorf("read invalid length (%d != %d)", n, nextSequenceLen)
		}

		component, err := NewKeyComponent(nextSequenceBuf)
		if err != nil {
			return Key{}, errors.Wrap(err, "error creating a key component")
		}

		components = append(components, component)
	}

	return NewKey(components...)
}

func (k Key) Append(component KeyComponent) Key {
	return Key{
		components: append(k.components, component),
	}
}

func (k Key) Len() int {
	return len(k.components)
}

func (k Key) Components() []KeyComponent {
	return k.components
}

func (k Key) Bytes() []byte {
	buf := bytes.NewBuffer(nil)

	for _, component := range k.components {
		buf.WriteByte(encodeComponentLength(component))
		buf.Write(component.b)
	}

	return buf.Bytes()
}

func (k Key) IsZero() bool {
	return len(k.components) == 0
}

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

func (b Bucket) DeleteBucket() error {
	if err := b.ForEach(func(item *badger.Item) error { // todo don't prefech values?
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

func (b Bucket) KeyInBucket(item *badger.Item) (Key, error) {
	itemKey, err := NewKeyFromBytes(item.KeyCopy(nil)) // todo copy only sometimes or later?
	if err != nil {
		return Key{}, errors.Wrap(err, "error parsing the key")
	}

	if len(itemKey.components) != len(b.prefix.components)+1 {
		return Key{}, errors.New("invalid item key length")
	}

	for i := range b.prefix.components {
		if !bytes.Equal(itemKey.components[i].b, b.prefix.components[i].b) {
			return Key{}, errors.New("invalid item key component")
		}
	}

	return NewKey(itemKey.components[len(b.prefix.components):]...)
}

func encodeComponentLength(component KeyComponent) byte {
	return byte(len(component.b))
}
