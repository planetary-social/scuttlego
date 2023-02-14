package known

type KnownMessageContent interface {
	Type() MessageContentType
}

type MessageContentType string // todo struct with strings.ToLower or pragma nocompare?

func (t MessageContentType) IsZero() bool {
	return t == ""
}
