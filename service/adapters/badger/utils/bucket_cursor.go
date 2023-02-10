package utils

import "github.com/dgraph-io/badger/v3"

type BucketIterator struct {
	bucket Bucket
	it     *badger.Iterator
}

func NewBucketIterator(bucket Bucket) *BucketIterator {
	return NewBucketIteratorWithModifiedOptions(bucket, nil)
}

func NewBucketIteratorWithModifiedOptions(bucket Bucket, fn func(options *badger.IteratorOptions)) *BucketIterator {
	options := badger.DefaultIteratorOptions
	options.Prefix = bucket.prefix.Bytes()
	options.PrefetchValues = false

	if fn != nil {
		fn(&options)
	}

	return &BucketIterator{
		bucket: bucket,
		it:     bucket.tx.NewIterator(options),
	}
}

func (i BucketIterator) Seek(key []byte) {
	targetKey, err := i.bucket.targetKey(key)
	if err != nil {
		panic(err)
	}

	i.it.Seek(targetKey)
}

func (i BucketIterator) Valid() bool {
	return i.it.Valid()
}

func (i BucketIterator) ValidForBucket() bool {
	return i.it.ValidForPrefix(i.bucket.prefix.Bytes())
}

func (i BucketIterator) Next() {
	i.it.Next()
}

func (i BucketIterator) Item() Item {
	return newItemAdapter(i.it.Item())
}

func (i BucketIterator) Close() {
	i.it.Close()
}

func (i BucketIterator) Rewind() {
	i.it.Rewind()
}
