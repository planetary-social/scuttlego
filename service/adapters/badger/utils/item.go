package utils

import "github.com/dgraph-io/badger/v3"

type Item interface {
	KeyCopy([]byte) []byte
	ValueCopy([]byte) ([]byte, error)

	DangerousKey(func(key []byte) error) error
	DangerousValue(func(value []byte) error) error
}

type itemAdapter struct {
	item *badger.Item
}

func newItemAdapter(item *badger.Item) itemAdapter {
	return itemAdapter{item: item}
}

func (i itemAdapter) KeyCopy(dst []byte) []byte {
	return i.item.KeyCopy(dst)
}

func (i itemAdapter) ValueCopy(dst []byte) ([]byte, error) {
	return i.item.ValueCopy(dst)
}

func (i itemAdapter) DangerousKey(f func(key []byte) error) error {
	return f(i.item.Key())
}

func (i itemAdapter) DangerousValue(f func(value []byte) error) error {
	return i.item.Value(f)
}
