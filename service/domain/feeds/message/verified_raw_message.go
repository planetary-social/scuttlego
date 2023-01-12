package message

import "github.com/boreq/errors"

type VerifiedRawMessage struct {
	data []byte
}

func NewVerifiedRawMessage(data []byte) (VerifiedRawMessage, error) {
	if len(data) == 0 {
		return VerifiedRawMessage{}, errors.New("empty slice")
	}

	tmp := make([]byte, len(data))
	copy(tmp, data)

	return VerifiedRawMessage{
		data: tmp,
	}, nil
}

func MustNewVerifiedRawMessage(data []byte) VerifiedRawMessage {
	v, err := NewVerifiedRawMessage(data)
	if err != nil {
		panic(err)
	}
	return v
}

func (m VerifiedRawMessage) Bytes() []byte {
	tmp := make([]byte, len(m.data))
	copy(tmp, m.data)
	return tmp
}

func (m VerifiedRawMessage) IsZero() bool {
	return len(m.data) == 0
}
