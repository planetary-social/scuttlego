package bolt

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"go.etcd.io/bbolt"
)

type WantListRepository struct {
	tx                  *bbolt.Tx
	currentTimeProvider commands.CurrentTimeProvider
	bucketPath          []utils.BucketName
}

func NewWantListRepository(
	tx *bbolt.Tx,
	currentTimeProvider commands.CurrentTimeProvider,
	bucketPath []utils.BucketName,
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

	value := bucket.Get(key)
	if value != nil {
		t, err := r.fromValue(value)
		if err != nil {
			return errors.Wrap(err, "failed to read the value")
		}

		if t.After(until) {
			return nil
		}
	}

	return bucket.Put(key, r.toValue(until))
}

func (r WantListRepository) Contains(id string) (bool, error) {
	bucket, err := r.getBucket()
	if err != nil {
		return false, errors.Wrap(err, "failed to get the bucket")
	}

	if bucket == nil {
		return false, nil
	}

	v := bucket.Get(r.toKey(id))
	if v == nil {
		return false, nil
	}

	until, err := r.fromValue(v)
	if err != nil {
		return false, errors.Wrap(err, "could not read the value")
	}

	now := r.currentTimeProvider.Get()

	if now.After(until) {
		return false, nil
	}

	return true, nil
}

func (r WantListRepository) Delete(id string) error {
	bucket, err := r.getBucket()
	if err != nil {
		return errors.Wrap(err, "failed to get the bucket")
	}

	if bucket == nil {
		return nil
	}

	if err := bucket.Delete(r.toKey(id)); err != nil {
		return errors.Wrap(err, "error calling delete")
	}

	return nil
}

func (r WantListRepository) List() ([]string, error) {
	var result []string
	var toDelete []string

	bucket, err := r.getBucket()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the bucket")
	}

	if bucket == nil {
		return nil, nil
	}

	now := r.currentTimeProvider.Get()

	if err := bucket.ForEach(func(k, v []byte) error {
		id := r.fromKey(k)

		until, err := r.fromValue(v)
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

func (r WantListRepository) createBucket() (*bbolt.Bucket, error) {
	return utils.CreateBucket(r.tx, r.bucketPath)
}

func (r WantListRepository) getBucket() (*bbolt.Bucket, error) {
	return utils.GetBucket(r.tx, r.bucketPath)
}
