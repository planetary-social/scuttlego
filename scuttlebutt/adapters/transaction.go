package adapters

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/scuttlebutt/commands"
	"go.etcd.io/bbolt"
)

type AdaptersFactory func(tx *bbolt.Tx) (commands.Adapters, error)

type TransactionProvider struct {
	db      *bbolt.DB
	factory AdaptersFactory
}

func NewTransactionProvider(db *bbolt.DB, factory AdaptersFactory) *TransactionProvider {
	return &TransactionProvider{db: db, factory: factory}
}

func (t TransactionProvider) Transact(f func(adapters commands.Adapters) error) error {
	return t.db.Update(func(tx *bbolt.Tx) error {
		adapters, err := t.factory(tx)
		if err != nil {
			return errors.Wrap(err, "failed to build adapters")
		}

		return f(adapters)
	})
}
