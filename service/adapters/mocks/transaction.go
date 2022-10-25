package mocks

import (
	"github.com/planetary-social/scuttlego/service/app/commands"
)

type MockTransactionProvider struct {
	adapters commands.Adapters
}

func NewMockTransactionProvider(adapters commands.Adapters) *MockTransactionProvider {
	return &MockTransactionProvider{adapters: adapters}
}

func (p *MockTransactionProvider) Transact(f func(adapters commands.Adapters) error) error {
	return f(p.adapters)
}
