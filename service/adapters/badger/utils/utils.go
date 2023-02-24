package utils

import (
	"bytes"
	"fmt"
	"math"

	"github.com/boreq/errors"
)

const maxKeyComponentLength = math.MaxUint8

type KeyComponent struct {
	b []byte
}

func NewKeyComponent(b []byte) (KeyComponent, error) {
	if len(b) == 0 {
		return KeyComponent{}, errors.New("empty key component")
	}

	if l := len(b); l > maxKeyComponentLength {
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
	return NewKeyFromBytesReusingComponents(b, nil)
}

func NewKeyFromBytesReusingComponents(b []byte, components []KeyComponent) (Key, error) {
	for {
		if len(b) == 0 {
			break
		}

		nextSequenceLen := int(b[0])
		b = b[1:]

		if len(b) < nextSequenceLen {
			return Key{}, errors.New("remaining data length < next sequence length")
		}

		nextSequence := b[0:nextSequenceLen]
		b = b[nextSequenceLen:]

		component, err := NewKeyComponent(nextSequence)
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

func encodeComponentLength(component KeyComponent) byte {
	return byte(len(component.b))
}
