package adapters

import (
	"github.com/boreq/errors"
	"go.etcd.io/bbolt"
)

type bucketName []byte

func getBucket(tx *bbolt.Tx, bucketNames []bucketName) *bbolt.Bucket {
	bucket := tx.Bucket(bucketNames[0])

	if bucket == nil {
		return nil
	}

	for i := 1; i < len(bucketNames); i++ {
		bucket = bucket.Bucket(bucketNames[i])
		if bucket == nil {
			return nil
		}
	}

	return bucket
}

func createBucket(tx *bbolt.Tx, bucketNames []bucketName) (bucket *bbolt.Bucket, err error) {
	bucket, err = tx.CreateBucketIfNotExists(bucketNames[0])
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
