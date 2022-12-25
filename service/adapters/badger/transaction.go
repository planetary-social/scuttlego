package badger

import (
	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/identity"
)

type AdaptersFactory func(tx *badger.Txn) (commands.Adapters, error)

type TransactionProvider struct {
	db      *badger.DB
	factory AdaptersFactory
}

func NewTransactionProvider(db *badger.DB, factory AdaptersFactory) *TransactionProvider {
	return &TransactionProvider{db: db, factory: factory}
}

func (t TransactionProvider) Transact(f func(adapters commands.Adapters) error) error {
	return t.db.Update(func(tx *badger.Txn) error {
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
	WantedFeedsRepository  *WantedFeedsRepository
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
