package badger

import (
	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/identity"
)

type CommandsAdaptersFactory func(tx *badger.Txn) (commands.Adapters, error)

type CommandsTransactionProvider struct {
	db      *badger.DB
	factory CommandsAdaptersFactory
}

func NewCommandsTransactionProvider(db *badger.DB, factory CommandsAdaptersFactory) *CommandsTransactionProvider {
	return &CommandsTransactionProvider{db: db, factory: factory}
}

func (t CommandsTransactionProvider) Transact(f func(adapters commands.Adapters) error) error {
	return t.db.Update(func(tx *badger.Txn) error {
		adapters, err := t.factory(tx)
		if err != nil {
			return errors.Wrap(err, "failed to build adapters")
		}

		return f(adapters)
	})
}

type QueriesAdaptersFactory func(tx *badger.Txn) (queries.Adapters, error)

type QueriesTransactionProvider struct {
	db      *badger.DB
	factory QueriesAdaptersFactory
}

func NewQueriesTransactionProvider(db *badger.DB, factory QueriesAdaptersFactory) *QueriesTransactionProvider {
	return &QueriesTransactionProvider{db: db, factory: factory}
}

func (t QueriesTransactionProvider) Transact(f func(adapters queries.Adapters) error) error {
	return t.db.Update(func(tx *badger.Txn) error { // todo https://github.com/planetary-social/scuttlego/issues/100
		adapters, err := t.factory(tx)
		if err != nil {
			return errors.Wrap(err, "failed to build adapters")
		}

		return f(adapters)
	})
}

type TestAdaptersFactory func(tx *badger.Txn, dependencies TestAdaptersDependencies) (TestAdapters, error)

type TestAdapters struct {
	BanListRepository      *BanListRepository
	BlobRepository         *BlobRepository
	BlobWantListRepository *BlobWantListRepository
	FeedWantListRepository *FeedWantListRepository
	MessageRepository      *MessageRepository
	ReceiveLogRepository   *ReceiveLogRepository
	SocialGraphRepository  *SocialGraphRepository
	PubRepository          *PubRepository
	FeedRepository         *FeedRepository
}

type TestAdaptersDependencies struct {
	BanListHasher        *mocks.BanListHasherMock
	CurrentTimeProvider  *mocks.CurrentTimeProviderMock
	RawMessageIdentifier *mocks.RawMessageIdentifierMock
	LocalIdentity        identity.Public
}

type TestTransactionProvider struct {
	db           *badger.DB
	factory      TestAdaptersFactory
	dependencies TestAdaptersDependencies
}

func NewTestTransactionProvider(db *badger.DB, dependencies TestAdaptersDependencies, factory TestAdaptersFactory) *TestTransactionProvider {
	return &TestTransactionProvider{db: db, factory: factory, dependencies: dependencies}
}

func (t TestTransactionProvider) Update(f func(adapters TestAdapters) error) error {
	return t.db.Update(func(tx *badger.Txn) error {
		adapters, err := t.factory(tx, t.dependencies)
		if err != nil {
			return errors.Wrap(err, "failed to build adapters")
		}

		return f(adapters)
	})
}
func (t TestTransactionProvider) View(f func(adapters TestAdapters) error) error {
	return t.db.View(func(tx *badger.Txn) error {
		adapters, err := t.factory(tx, t.dependencies)
		if err != nil {
			return errors.Wrap(err, "failed to build adapters")
		}

		return f(adapters)
	})
}
