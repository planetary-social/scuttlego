package mocks

import (
	"sync"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
)

type MockCloserStream struct {
	writtenMessages []MockCloserStreamWriteMessageCall
	writtenErrors   []error
	lock            sync.Mutex // locks writtenMessages and writtenErrors
}

func NewMockCloserStream() *MockCloserStream {
	return &MockCloserStream{}
}

func (m *MockCloserStream) WriteMessage(body []byte, bodyType transport.MessageBodyType) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	cpy := make([]byte, len(body))
	copy(cpy, body)
	m.writtenMessages = append(m.writtenMessages, MockCloserStreamWriteMessageCall{
		Body:     cpy,
		BodyType: bodyType,
	})
	return nil
}

func (m *MockCloserStream) WrittenMessages() []MockCloserStreamWriteMessageCall {
	m.lock.Lock()
	defer m.lock.Unlock()

	tmp := make([]MockCloserStreamWriteMessageCall, len(m.writtenMessages))
	copy(tmp, m.writtenMessages)
	return tmp
}

func (m *MockCloserStream) CloseWithError(err error) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.writtenErrors = append(m.writtenErrors, err)
	return nil
}

func (m *MockCloserStream) WrittenErrors() []error {
	m.lock.Lock()
	defer m.lock.Unlock()

	tmp := make([]error, len(m.writtenErrors))
	copy(tmp, m.writtenErrors)
	return tmp
}

func (m *MockCloserStream) IncomingMessages() (<-chan rpc.IncomingMessage, error) {
	return nil, errors.New("not implemented")
}

type MockCloserStreamWriteMessageCall struct {
	Body     []byte
	BodyType transport.MessageBodyType
}
