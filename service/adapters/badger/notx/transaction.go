package notx

import (
	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	badgeradapters "github.com/planetary-social/scuttlego/service/adapters/badger"
)

type TxAdaptersFactory func(tx *badger.Txn) (TxAdapters, error)

type TxAdapters struct {
	BanListRepository      *badgeradapters.BanListRepository
	BlobRepository         *badgeradapters.BlobRepository
	BlobWantListRepository *badgeradapters.BlobWantListRepository
	FeedWantListRepository *badgeradapters.FeedWantListRepository
	MessageRepository      *badgeradapters.MessageRepository
	ReceiveLogRepository   *badgeradapters.ReceiveLogRepository
	SocialGraphRepository  *badgeradapters.SocialGraphRepository
	PubRepository          *badgeradapters.PubRepository
}

type TransactionProvider struct {
	db      *badger.DB
	factory TxAdaptersFactory
}

func NewTransactionProvider(db *badger.DB, factory TxAdaptersFactory) *TransactionProvider {
	return &TransactionProvider{db: db, factory: factory}
}

func (t TransactionProvider) Update(f func(adapters TxAdapters) error) error {
	return t.db.Update(func(tx *badger.Txn) error {
		adapters, err := t.factory(tx)
		if err != nil {
			return errors.Wrap(err, "failed to build adapters")
		}

		return f(adapters)
	})
}
func (t TransactionProvider) View(f func(adapters TxAdapters) error) error {
	return t.db.View(func(tx *badger.Txn) error {
		adapters, err := t.factory(tx)
		if err != nil {
			return errors.Wrap(err, "failed to build adapters")
		}

		return f(adapters)
	})
}

type TestAdapters struct {
	ReadBlobWantListRepository *NoTxBlobWantListRepository
}
