package badger

import (
	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/app/commands"
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

type TestAdaptersFactory func(tx *badger.Txn) (TestAdapters, error)

type TestAdapters struct {
	BanList               *BanListRepository
	BlobRepository        *BlobRepository
	BlobWantList          *BlobWantListRepository
	FeedWantList          *FeedWantListRepository
	MessageRepository     *MessageRepository
	ReceiveLog            *ReceiveLogRepository
	SocialGraphRepository *SocialGraphRepository

	// todo name either all ...repository or strip it from all names

	BanListHasher       *mocks.BanListHasherMock
	CurrentTimeProvider *mocks.CurrentTimeProviderMock
}

type TestTransactionProvider struct {
	db      *badger.DB
	factory TestAdaptersFactory
}

func NewTestTransactionProvider(db *badger.DB, factory TestAdaptersFactory) *TestTransactionProvider {
	return &TestTransactionProvider{db: db, factory: factory}
}

func (t TestTransactionProvider) Update(f func(adapters TestAdapters) error) error {
	return t.db.Update(func(tx *badger.Txn) error {
		adapters, err := t.factory(tx)
		if err != nil {
			return errors.Wrap(err, "failed to build adapters")
		}

		return f(adapters)
	})
}
func (t TestTransactionProvider) View(f func(adapters TestAdapters) error) error {
	return t.db.View(func(tx *badger.Txn) error {
		adapters, err := t.factory(tx)
		if err != nil {
			return errors.Wrap(err, "failed to build adapters")
		}

		return f(adapters)
	})
}
