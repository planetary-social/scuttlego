package notx

import (
	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	badgeradapters "github.com/planetary-social/scuttlego/service/adapters/badger"
)

type TestAdapters struct {
	NoTxBlobWantListRepository *NoTxBlobWantListRepository
	NoTxWantedFeedsRepository  *NoTxWantedFeedsRepository
}

type TxAdapters struct {
	BanListRepository      *badgeradapters.BanListRepository
	BlobRepository         *badgeradapters.BlobRepository
	BlobWantListRepository *badgeradapters.BlobWantListRepository
	FeedWantListRepository *badgeradapters.FeedWantListRepository
	MessageRepository      *badgeradapters.MessageRepository
	ReceiveLogRepository   *badgeradapters.ReceiveLogRepository
	SocialGraphRepository  *badgeradapters.SocialGraphRepository
	PubRepository          *badgeradapters.PubRepository
	FeedRepository         *badgeradapters.FeedRepository
	WantedFeedsRepository  *badgeradapters.WantedFeedsRepository
}

type TransactionProvider interface {
	Update(f func(adapters TxAdapters) error) error
	View(f func(adapters TxAdapters) error) error
}

type TxAdaptersFactory func(tx *badger.Txn) (TxAdapters, error)

type TxAdaptersFactoryTransactionProvider struct {
	db      *badger.DB
	factory TxAdaptersFactory
}

func NewTxAdaptersFactoryTransactionProvider(db *badger.DB, factory TxAdaptersFactory) *TxAdaptersFactoryTransactionProvider {
	return &TxAdaptersFactoryTransactionProvider{db: db, factory: factory}
}

func (t TxAdaptersFactoryTransactionProvider) Update(f func(adapters TxAdapters) error) error {
	return t.db.Update(func(tx *badger.Txn) error {
		adapters, err := t.factory(tx)
		if err != nil {
			return errors.Wrap(err, "failed to build adapters")
		}

		return f(adapters)
	})
}
func (t TxAdaptersFactoryTransactionProvider) View(f func(adapters TxAdapters) error) error {
	return t.db.View(func(tx *badger.Txn) error {
		adapters, err := t.factory(tx)
		if err != nil {
			return errors.Wrap(err, "failed to build adapters")
		}

		return f(adapters)
	})
}

type TestTxAdaptersFactory func(tx *badger.Txn, dependencies badgeradapters.TestAdaptersDependencies) (TxAdapters, error)

type TestTxAdaptersFactoryTransactionProvider struct {
	db           *badger.DB
	factory      TestTxAdaptersFactory
	dependencies badgeradapters.TestAdaptersDependencies
}

func NewTestTxAdaptersFactoryTransactionProvider(db *badger.DB, factory TestTxAdaptersFactory, dependencies badgeradapters.TestAdaptersDependencies) *TestTxAdaptersFactoryTransactionProvider {
	return &TestTxAdaptersFactoryTransactionProvider{db: db, factory: factory, dependencies: dependencies}
}

func (t TestTxAdaptersFactoryTransactionProvider) Update(f func(adapters TxAdapters) error) error {
	return t.db.Update(func(tx *badger.Txn) error {
		adapters, err := t.factory(tx, t.dependencies)
		if err != nil {
			return errors.Wrap(err, "failed to build adapters")
		}

		return f(adapters)
	})
}
func (t TestTxAdaptersFactoryTransactionProvider) View(f func(adapters TxAdapters) error) error {
	return t.db.View(func(tx *badger.Txn) error {
		adapters, err := t.factory(tx, t.dependencies)
		if err != nil {
			return errors.Wrap(err, "failed to build adapters")
		}

		return f(adapters)
	})
}
