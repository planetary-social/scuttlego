package refs

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/boreq/errors"
	"strings"
)

type Message struct {
	s string
	b []byte
}

const (
	messagePrefix = "%"
	messageSuffix = ".sha256"

	messageHashLength = 32
)

func NewMessage(s string) (Message, error) {
	if !strings.HasPrefix(s, messagePrefix) {
		return Message{}, errors.New("invalid prefix")
	}

	if !strings.HasSuffix(s, messageSuffix) {
		return Message{}, errors.New("invalid suffix")
	}

	noSuffixAndPrefix := s[len(messagePrefix) : len(s)-len(messageSuffix)]

	b, err := base64.StdEncoding.DecodeString(noSuffixAndPrefix)
	if err != nil {
		return Message{}, errors.Wrapf(err, "invalid base64 '%s'", noSuffixAndPrefix)
	}

	if l := len(b); l != messageHashLength {
		return Message{}, fmt.Errorf("invalid hash length '%d'", l)
	}

	return Message{s, b}, nil
}

func MustNewMessage(s string) Message {
	r, err := NewMessage(s)
	if err != nil {
		panic(err)
	}
	return r
}

func (m Message) Bytes() []byte {
	return m.b
}

func (m Message) String() string {
	return m.s
}

func (m Message) IsZero() bool {
	return len(m.b) == 0
}

func (m Message) Equal(o Message) bool {
	return bytes.Equal(m.b, o.b)
}
