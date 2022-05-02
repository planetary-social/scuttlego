package message

import "github.com/boreq/errors"

type RawMessage struct {
	data []byte
}

func NewRawMessage(data []byte) (RawMessage, error) {
	if len(data) == 0 {
		return RawMessage{}, errors.New("empty slice")
	}

	tmp := make([]byte, len(data))
	copy(tmp, data)

	return RawMessage{
		data: tmp,
	}, nil
}

func MustNewRawMessage(data []byte) RawMessage {
	v, err := NewRawMessage(data)
	if err != nil {
		panic(err)
	}
	return v
}

func (m RawMessage) Bytes() []byte {
	tmp := make([]byte, len(m.data))
	copy(tmp, m.data)
	return tmp
}

func (m RawMessage) IsZero() bool {
	return len(m.data) == 0
}
