package bolt

import (
	"github.com/boreq/errors"
	"go.etcd.io/bbolt"
)

type bucketName []byte

func getBucket(tx *bbolt.Tx, bucketNames []bucketName) (*bbolt.Bucket, error) {
	if len(bucketNames) == 0 {
		return nil, errors.New("empty bucket names")
	}

	bucket := tx.Bucket(bucketNames[0])

	if bucket == nil {
		return nil, nil
	}

	for i := 1; i < len(bucketNames); i++ {
		bucket = bucket.Bucket(bucketNames[i])
		if bucket == nil {
			return nil, nil
		}
	}

	return bucket, nil
}

func createBucket(tx *bbolt.Tx, bucketNames []bucketName) (*bbolt.Bucket, error) {
	if len(bucketNames) == 0 {
		return nil, errors.New("empty bucket names")
	}

	bucket, err := tx.CreateBucketIfNotExists(bucketNames[0])
	if err != nil {
		return nil, errors.Wrap(err, "could not create a bucket")
	}

	for i := 1; i < len(bucketNames); i++ {
		bucket, err = bucket.CreateBucketIfNotExists(bucketNames[i])
		if err != nil {
			return nil, errors.Wrap(err, "could not create a bucket")
		}
	}

	return bucket, nil
}
