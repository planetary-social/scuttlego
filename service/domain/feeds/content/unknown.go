package content

type Unknown struct {
	b []byte
}

func NewUnknown(b []byte) (Unknown, error) {
	return Unknown{b: b}, nil
}

func MustNewUnknown(b []byte) Unknown {
	u, err := NewUnknown(b)
	if err != nil {
		panic(err)
	}
	return u
}

func (u Unknown) Type() MessageContentType {
	return ""
}

func (u Unknown) Bytes() []byte {
	tmp := make([]byte, len(u.b))
	copy(tmp, u.b)
	return tmp
}
