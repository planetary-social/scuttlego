package message

import "errors"

type RawContent struct {
	data []byte
}

func NewRawContent(data []byte) (RawContent, error) {
	if len(data) == 0 {
		return RawContent{}, errors.New("empty content")
	}

	tmp := make([]byte, len(data))
	copy(tmp, data)

	return RawContent{
		data: tmp,
	}, nil
}

func MustNewRawContent(data []byte) RawContent {
	v, err := NewRawContent(data)
	if err != nil {
		panic(err)
	}
	return v
}

func (m RawContent) Bytes() []byte {
	tmp := make([]byte, len(m.data))
	copy(tmp, m.data)
	return tmp
}

func (m RawContent) IsZero() bool {
	return len(m.data) == 0
}
