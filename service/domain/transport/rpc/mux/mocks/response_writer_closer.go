package mocks

type MockResponseWriterCloser struct {
	WrittenMessages [][]byte
	WrittenErrors   []error
}

func NewMockResponseWriterCloser() *MockResponseWriterCloser {
	return &MockResponseWriterCloser{}
}

func (m *MockResponseWriterCloser) WriteMessage(body []byte) error {
	cpy := make([]byte, len(body))
	copy(cpy, body)
	m.WrittenMessages = append(m.WrittenMessages, cpy)
	return nil
}

func (m *MockResponseWriterCloser) CloseWithError(err error) error {
	m.WrittenErrors = append(m.WrittenErrors, err)
	return nil
}
