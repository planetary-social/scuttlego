package migrations

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/planetary-social/scuttlego/migrations"
	"github.com/planetary-social/scuttlego/service/adapters/badger/utils"
)

type BadgerStorage struct {
	db *badger.DB
}

func NewBadgerStorage(db *badger.DB) *BadgerStorage {
	return &BadgerStorage{db: db}
}

func (b *BadgerStorage) LoadState(name string) (migrations.State, error) {
	var state migrations.State

	if err := b.db.View(func(txn *badger.Txn) error {
		bucket := b.stateBucket(txn)

		item, err := bucket.Get([]byte(name))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return migrations.ErrStateNotFound
			}
			return errors.Wrap(err, "could not get the item")
		}

		if err := item.Value(func(val []byte) error {
			if err := json.Unmarshal(val, &state); err != nil {
				return errors.Wrap(err, "error unmarshaling state")
			}
			return nil
		}); err != nil {
			return errors.Wrap(err, "error getting value")
		}

		return nil
	}); err != nil {
		if errors.Is(err, migrations.ErrStateNotFound) {
			return nil, err
		}
		return nil, errors.Wrap(err, "transaction failed")
	}

	return state, nil
}

func (b *BadgerStorage) SaveState(name string, state migrations.State) error {
	marshaledState, err := json.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "error marshaling state")
	}

	if err := b.db.Update(func(txn *badger.Txn) error {
		bucket := b.stateBucket(txn)
		return bucket.Set([]byte(name), marshaledState)
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}
	return nil
}

func (b *BadgerStorage) LoadStatus(name string) (migrations.Status, error) {
	var status migrations.Status

	if err := b.db.View(func(txn *badger.Txn) error {
		bucket := b.statusBucket(txn)

		item, err := bucket.Get([]byte(name))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return migrations.ErrStatusNotFound
			}
			return errors.Wrap(err, "could not get the item")
		}

		if err := item.Value(func(val []byte) error {
			tmp, err := unmarshalStatus(string(val))
			if err != nil {
				return errors.Wrap(err, "error unmarshaling status")
			}
			status = tmp
			return nil
		}); err != nil {
			return errors.Wrap(err, "error getting value")
		}

		return nil
	}); err != nil {
		if errors.Is(err, migrations.ErrStatusNotFound) {
			return migrations.Status{}, err
		}
		return migrations.Status{}, errors.Wrap(err, "transaction failed")
	}

	return status, nil
}

func (b *BadgerStorage) SaveStatus(name string, status migrations.Status) error {
	marshaledStatus, err := marshalStatus(status)
	if err != nil {
		return errors.Wrap(err, "error marshaling status")
	}

	if err := b.db.Update(func(txn *badger.Txn) error {
		bucket := b.statusBucket(txn)
		return bucket.Set([]byte(name), []byte(marshaledStatus))
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}
	return nil
}

func (b *BadgerStorage) statusBucket(tx *badger.Txn) utils.Bucket {
	return utils.MustNewBucket(tx, utils.MustNewKey(
		utils.MustNewKeyComponent([]byte(migrationsBucket)),
		utils.MustNewKeyComponent([]byte(migrationsBucketStatus)),
	))
}

func (b *BadgerStorage) stateBucket(tx *badger.Txn) utils.Bucket {
	return utils.MustNewBucket(tx, utils.MustNewKey(
		utils.MustNewKeyComponent([]byte(migrationsBucket)),
		utils.MustNewKeyComponent([]byte(migrationsBucketState)),
	))
}