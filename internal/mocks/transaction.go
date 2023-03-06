package mocks

import (
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
)

type MockCommandsTransactionProvider struct {
	adapters commands.Adapters
}

func NewMockCommandsTransactionProvider(adapters commands.Adapters) *MockCommandsTransactionProvider {
	return &MockCommandsTransactionProvider{adapters: adapters}
}

func (p *MockCommandsTransactionProvider) Transact(f func(adapters commands.Adapters) error) error {
	return f(p.adapters)
}

type MockQueriesTransactionProvider struct {
	adapters queries.Adapters
}

func NewMockQueriesTransactionProvider(adapters queries.Adapters) *MockQueriesTransactionProvider {
	return &MockQueriesTransactionProvider{adapters: adapters}
}

func (p *MockQueriesTransactionProvider) Transact(f func(adapters queries.Adapters) error) error {
	return f(p.adapters)
}
