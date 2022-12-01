package migrations

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/migrations"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"go.etcd.io/bbolt"
)

const (
	migrationsBucket       = "migrations"
	migrationsBucketStatus = "status"
	migrationsBucketState  = "state"
)

type BoltStorage struct {
	db *bbolt.DB
}

func NewBoltStorage(db *bbolt.DB) *BoltStorage {
	return &BoltStorage{db: db}
}

func (b *BoltStorage) LoadState(name string) (migrations.State, error) {
	return migrations.State{}, migrations.ErrStateNotFound
}

func (b *BoltStorage) SaveState(name string, state migrations.State) error {
	return nil
}

func (b *BoltStorage) LoadStatus(name string) (migrations.Status, error) {
	var status migrations.Status

	if err := b.db.View(func(tx *bbolt.Tx) error {
		b, err := utils.GetBucket(tx, b.statusBucket())
		if err != nil {
			return errors.Wrap(err, "error creating bucket")
		}

		if b == nil {
			return migrations.ErrStatusNotFound
		}

		v := b.Get([]byte(name))
		if v == nil {
			return migrations.ErrStatusNotFound
		}

		status, err = unmarshalStatus(string(v))
		if err != nil {
			return errors.Wrap(err, "error unmarshaling status")
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

func (b *BoltStorage) SaveStatus(name string, status migrations.Status) error {
	marshaledStatus, err := marshalStatus(status)
	if err != nil {
		return errors.Wrap(err, "error marshaling status")
	}

	if err := b.db.Batch(func(tx *bbolt.Tx) error {
		b, err := utils.CreateBucket(tx, b.statusBucket())
		if err != nil {
			return errors.Wrap(err, "error creating bucket")
		}

		return b.Put([]byte(name), []byte(marshaledStatus))
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}
	return nil
}

func (b *BoltStorage) statusBucket() []utils.BucketName {
	return []utils.BucketName{
		utils.BucketName(migrationsBucket),
		utils.BucketName(migrationsBucketStatus),
	}
}

func (b *BoltStorage) stateBucket() []utils.BucketName {
	return []utils.BucketName{
		utils.BucketName(migrationsBucket),
		utils.BucketName(migrationsBucketState),
	}
}

const (
	statusFailed   = "failed"
	statusFinished = "finished"
)

func marshalStatus(status migrations.Status) (string, error) {
	switch status {
	case migrations.StatusFailed:
		return statusFailed, nil
	case migrations.StatusFinished:
		return statusFinished, nil
	default:
		return "", errors.New("unknown status")
	}
}

func unmarshalStatus(status string) (migrations.Status, error) {
	switch status {
	case statusFailed:
		return migrations.StatusFailed, nil
	case statusFinished:
		return migrations.StatusFinished, nil
	default:
		return migrations.Status{}, errors.New("unknown status")
	}
}
