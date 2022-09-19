package mocks

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type MockCloserStream struct {
	WrittenMessages [][]byte
	WrittenErrors   []error
}

func NewMockCloserStream() *MockCloserStream {
	return &MockCloserStream{}
}

func (m *MockCloserStream) WriteMessage(body []byte) error {
	cpy := make([]byte, len(body))
	copy(cpy, body)
	m.WrittenMessages = append(m.WrittenMessages, cpy)
	return nil
}

func (m *MockCloserStream) CloseWithError(err error) error {
	m.WrittenErrors = append(m.WrittenErrors, err)
	return nil
}

func (m *MockCloserStream) IncomingMessages() (<-chan rpc.IncomingMessage, error) {
	return nil, errors.New("not implemented")
}
