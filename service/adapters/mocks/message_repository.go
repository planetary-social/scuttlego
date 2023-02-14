package mocks

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type MessageRepositoryMock struct {
	CountReturnValue int

	getCalls        []MessageRepositoryMockGetCall
	getReturnValues map[string]message.Message
}

func NewMessageRepositoryMock() *MessageRepositoryMock {
	return &MessageRepositoryMock{
		getReturnValues: make(map[string]message.Message),
	}
}

func (m *MessageRepositoryMock) Count() (int, error) {
	return m.CountReturnValue, nil
}

func (m *MessageRepositoryMock) MockGet(msg message.Message) {
	_, ok := m.getReturnValues[msg.Id().String()]
	if ok {
		panic("message with this id was already mocked")
	}
	m.getReturnValues[msg.Id().String()] = msg
}

func (m *MessageRepositoryMock) Get(id refs.Message) (message.Message, error) {
	m.getCalls = append(m.getCalls, MessageRepositoryMockGetCall{Id: id})
	msg, ok := m.getReturnValues[id.String()]
	if !ok {
		return message.Message{}, errors.New("not mocked")
	}
	return msg, nil
}

func (m *MessageRepositoryMock) GetCalls() []MessageRepositoryMockGetCall {
	return m.getCalls
}

type MessageRepositoryMockGetCall struct {
	Id refs.Message
}
