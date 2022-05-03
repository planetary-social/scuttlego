package message

import "errors"

type RawMessageContent struct {
	data []byte
}

func NewRawMessageContent(data []byte) (RawMessageContent, error) {
	if len(data) == 0 {
		return RawMessageContent{}, errors.New("empty content")
	}

	tmp := make([]byte, len(data))
	copy(tmp, data)

	return RawMessageContent{
		data: tmp,
	}, nil
}

func MustNewRawMessageContent(data []byte) RawMessageContent {
	v, err := NewRawMessageContent(data)
	if err != nil {
		panic(err)
	}
	return v
}

func (m RawMessageContent) Bytes() []byte {
	tmp := make([]byte, len(m.data))
	copy(tmp, m.data)
	return tmp
}

func (m RawMessageContent) IsZero() bool {
	return len(m.data) == 0
}
