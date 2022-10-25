package bolt

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"go.etcd.io/bbolt"
)

var bucketBlobWantList = utils.BucketName("blob_want_list")

type BlobWantListRepository struct {
	tx                  *bbolt.Tx
	currentTimeProvider commands.CurrentTimeProvider
}

func NewBlobWantListRepository(
	tx *bbolt.Tx,
	currentTimeProvider commands.CurrentTimeProvider,
) *BlobWantListRepository {
	return &BlobWantListRepository{
		tx:                  tx,
		currentTimeProvider: currentTimeProvider,
	}
}

func (r BlobWantListRepository) Add(id refs.Blob, until time.Time) error {
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

func (r BlobWantListRepository) Contains(id refs.Blob) (bool, error) {
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

func (r BlobWantListRepository) Delete(id refs.Blob) error {
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

func (r BlobWantListRepository) List() ([]refs.Blob, error) {
	var result []refs.Blob
	var toDelete []refs.Blob

	bucket, err := r.getBucket()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the bucket")
	}

	if bucket == nil {
		return nil, nil
	}

	now := r.currentTimeProvider.Get()

	if err := bucket.ForEach(func(k, v []byte) error {
		id, err := r.fromKey(k)
		if err != nil {
			return errors.Wrap(err, "could not read the key")
		}

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

func (r BlobWantListRepository) toKey(id refs.Blob) []byte {
	return []byte(id.String())
}

func (r BlobWantListRepository) fromKey(key []byte) (refs.Blob, error) {
	return refs.NewBlob(string(key))
}

func (r BlobWantListRepository) toValue(t time.Time) []byte {
	return []byte(t.Format(time.RFC3339))
}

func (r BlobWantListRepository) fromValue(v []byte) (time.Time, error) {
	return time.Parse(time.RFC3339, string(v))
}

func (r BlobWantListRepository) createBucket() (*bbolt.Bucket, error) {
	return utils.CreateBucket(r.tx, r.bucketPath())
}

func (r BlobWantListRepository) getBucket() (*bbolt.Bucket, error) {
	return utils.GetBucket(r.tx, r.bucketPath())
}

func (r BlobWantListRepository) bucketPath() []utils.BucketName {
	return []utils.BucketName{
		bucketBlobWantList,
	}
}
