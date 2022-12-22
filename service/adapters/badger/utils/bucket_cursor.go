package utils

import "github.com/dgraph-io/badger/v3"

type BucketIterator struct {
	bucket Bucket
	it     *badger.Iterator
}

func NewBucketIterator(bucket Bucket) *BucketIterator {
	return &BucketIterator{
		bucket: bucket,
		it:     bucket.tx.NewIterator(badger.DefaultIteratorOptions),
	}
}

func (i BucketIterator) Seek(key []byte) {
	targetKey, err := i.bucket.targetKey(key)
	if err != nil {
		panic(err)
	}

	i.it.Seek(targetKey)
}

func (i BucketIterator) ValidForBucket() bool {
	return i.it.ValidForPrefix(i.bucket.prefix.Bytes())
}

func (i BucketIterator) Next() {
	i.it.Next()
}

func (i BucketIterator) Item() *badger.Item {
	return i.it.Item()
}

func (i BucketIterator) Close() {
	i.it.Close()
}
