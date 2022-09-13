package utils

import (
	"github.com/boreq/errors"
	"go.etcd.io/bbolt"
)

type BucketName []byte

func GetBucket(tx *bbolt.Tx, bucketNames []BucketName) (*bbolt.Bucket, error) {
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

func CreateBucket(tx *bbolt.Tx, bucketNames []BucketName) (*bbolt.Bucket, error) {
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

func DeleteBucket(tx *bbolt.Tx, parentBucketNames []BucketName, bucketName BucketName) error {
	bucket, err := GetBucket(tx, parentBucketNames)
	if err != nil {
		return errors.Wrap(err, "failed to get the parent bucket")
	}

	if bucket == nil {
		return nil
	}

	return bucket.DeleteBucket(bucketName)
}

func BucketIsEmpty(bucket *bbolt.Bucket) bool {
	c := bucket.Cursor()

	if k, _ := c.First(); k != nil {
		return false
	}

	return true
}
