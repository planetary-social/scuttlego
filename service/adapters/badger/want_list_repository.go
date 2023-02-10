package badger

import (
	"time"

	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/planetary-social/scuttlego/service/adapters/badger/utils"
	"github.com/planetary-social/scuttlego/service/app/commands"
)

type WantListRepository struct {
	tx                  *badger.Txn
	currentTimeProvider commands.CurrentTimeProvider
	bucketPath          utils.Key
}

func NewWantListRepository(
	tx *badger.Txn,
	currentTimeProvider commands.CurrentTimeProvider,
	bucketPath utils.Key,
) *WantListRepository {
	return &WantListRepository{
		tx:                  tx,
		currentTimeProvider: currentTimeProvider,
		bucketPath:          bucketPath,
	}
}

func (r WantListRepository) Add(id string, until time.Time) error {
	bucket, err := r.createBucket()
	if err != nil {
		return errors.Wrap(err, "failed to get the bucket")
	}

	key := r.toKey(id)

	shouldSet, err := r.shouldSet(bucket, key, until)
	if err != nil {
		return errors.Wrap(err, "should set error")
	}

	if !shouldSet {
		return nil
	}

	return bucket.Set(key, r.toValue(until))
}

func (r WantListRepository) shouldSet(bucket utils.Bucket, key []byte, until time.Time) (bool, error) {
	item, err := bucket.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return true, nil
		}
		return false, errors.Wrap(err, "error getting the existing item")
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return false, errors.Wrap(err, "error getting the value")
	}

	t, err := r.fromValue(value)
	if err != nil {
		return false, errors.Wrap(err, "failed to convert the value")
	}

	shouldSet := !t.After(until)

	return shouldSet, nil
}

func (r WantListRepository) Contains(id string) (bool, error) {
	bucket, err := r.createBucket()
	if err != nil {
		return false, errors.Wrap(err, "failed to get the bucket")
	}

	item, err := bucket.Get(r.toKey(id))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return false, nil
		}
		return false, errors.Wrap(err, "error getting the value")
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return false, errors.Wrap(err, "error getting the value")
	}

	until, err := r.fromValue(value)
	if err != nil {
		return false, errors.Wrap(err, "could not convert the value")
	}

	now := r.currentTimeProvider.Get()
	valueIsStillValid := !now.After(until)

	return valueIsStillValid, nil
}

func (r WantListRepository) Delete(id string) error {
	bucket, err := r.createBucket()
	if err != nil {
		return errors.Wrap(err, "failed to get the bucket")
	}

	if err := bucket.Delete(r.toKey(id)); err != nil {
		return errors.Wrap(err, "delete error")
	}

	return nil
}

func (r WantListRepository) List() ([]string, error) {
	var result []string
	var toDelete []string

	bucket, err := r.createBucket()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the bucket")
	}

	now := r.currentTimeProvider.Get()

	if err := bucket.ForEach(func(item utils.Item) error {
		keyInBucket, err := bucket.KeyInBucket(item)
		if err != nil {
			return errors.Wrap(err, "error determining key in bucket")
		}

		id := r.fromKey(keyInBucket.Bytes())

		val, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Wrap(err, "could not get the value")
		}

		until, err := r.fromValue(val)
		if err != nil {
			return errors.Wrap(err, "could not read the value")
		}

		if now.After(until) {
			toDelete = append(toDelete, id)
			return nil
		}

		result = append(result, id)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "for each failed")
	}

	for _, id := range toDelete {
		if err := bucket.Delete(r.toKey(id)); err != nil {
			return nil, errors.Wrap(err, "deletion failed")
		}
	}

	return result, nil
}

func (r WantListRepository) toKey(id string) []byte {
	return []byte(id)
}

func (r WantListRepository) fromKey(key []byte) string {
	return string(key)
}

func (r WantListRepository) toValue(t time.Time) []byte {
	return []byte(t.Format(time.RFC3339))
}

func (r WantListRepository) fromValue(v []byte) (time.Time, error) {
	return time.Parse(time.RFC3339, string(v))
}

func (r WantListRepository) createBucket() (utils.Bucket, error) {
	return utils.NewBucket(r.tx, r.bucketPath)
}
