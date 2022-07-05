package mocks

type MessageRepositoryMock struct {
	CountReturnValue int
}

func NewMessageRepositoryMock() *MessageRepositoryMock {
	return &MessageRepositoryMock{}
}

func (m MessageRepositoryMock) Count() (int, error) {
	return m.CountReturnValue, nil
}
