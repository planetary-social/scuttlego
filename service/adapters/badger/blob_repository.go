package badger

import (
	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/planetary-social/scuttlego/service/adapters/badger/utils"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

var (
	bucketBlobsKeyComponent          = utils.MustNewKeyComponent([]byte("blobs"))
	bucketBlobsByMessageKeyComponent = utils.MustNewKeyComponent([]byte("by_message"))
	bucketBlobsByBlobKeyComponent    = utils.MustNewKeyComponent([]byte("by_blob"))
)

type BlobRepository struct {
	tx *badger.Txn
}

func NewBlobRepository(tx *badger.Txn) *BlobRepository {
	return &BlobRepository{
		tx: tx,
	}
}

func (r BlobRepository) Put(msgRef refs.Message, blob feeds.BlobToSave) error {
	byMessageBucket, err := r.createBucketByMessageBlobRefs(msgRef)
	if err != nil {
		return errors.Wrap(err, "could not create by_message bucket")
	}

	for _, blobRef := range blob.Blobs() {
		if err := byMessageBucket.Set([]byte(blobRef.String()), nil); err != nil {
			return errors.Wrap(err, "by_message bucket put failed")
		}

		byBlobBucket, err := r.createBucketByBlobMessageRefs(blobRef)
		if err != nil {
			return errors.Wrap(err, "could not create by_blob bucket")
		}

		if err := byBlobBucket.Set([]byte(msgRef.String()), nil); err != nil {
			return errors.Wrap(err, "by_blob bucket put failed")
		}
	}

	return nil
}

func (r BlobRepository) Delete(msgRef refs.Message) error {
	byMessageBlobRefsBucket, err := r.createBucketByMessageBlobRefs(msgRef)
	if err != nil {
		return errors.Wrap(err, "could not get blob refs bucket")
	}

	//byBlobBucket, err := r.createBucketByBlob()
	//if err != nil {
	//	return errors.Wrap(err, "could not get the by blob bucket")
	//}

	if err := byMessageBlobRefsBucket.ForEach(func(item *badger.Item) error {
		k, err := byMessageBlobRefsBucket.KeyInBucket(item)
		if err != nil {
			return errors.Wrap(err, "error determining item key in bucket")
		}

		if k.Len() != 1 {
			return errors.New("invalid key length")
		}

		blobRef, err := refs.NewBlob(string(k.Components()[0].Bytes()))
		if err != nil {
			return errors.Wrap(err, "new blob ref error")
		}

		byBlobMessageRefsBucket, err := r.createBucketByBlobMessageRefs(blobRef)
		if err != nil {
			return errors.Wrap(err, "could not get message refs bucket")
		}

		if err := byBlobMessageRefsBucket.Delete([]byte(msgRef.String())); err != nil {
			return errors.Wrap(err, "delete error")
		}

		//blobRefBucket := byBlobBucket.Bucket(k)
		//
		//if err := blobRefBucket.Delete([]byte(msgRef.String())); err != nil {
		//	return errors.Wrap(err, "could not remove from the bucket")
		//}
		//
		//if utils.BucketIsEmpty(blobRefBucket) {
		//	if err := byBlobBucket.DeleteBucket(k); err != nil {
		//		return errors.Wrap(err, "could not remove the blob ref bucket")
		//	}
		//}

		return nil
	}); err != nil {
		return errors.Wrap(err, "foreach error")
	}

	//c := byMessageBlobRefsBucket.Cursor()
	//for k, _ := c.First(); k != nil; k, _ = c.Next() {
	//}

	//if err := utils.DeleteBucket(r.tx, r.bucketPathByMessage(), utils.BucketName(msgRef.String())); err != nil {
	//	return errors.Wrap(err, "could not delete the message bucket")
	//}

	if err := byMessageBlobRefsBucket.DeleteBucket(); err != nil {
		return errors.Wrap(err, "error deleting the blob refs bucket")
	}

	return nil
}

func (r BlobRepository) ListBlobs(msgRef refs.Message) ([]refs.Blob, error) {
	bucket, err := r.createBucketByMessageBlobRefs(msgRef)
	if err != nil {
		return nil, errors.Wrap(err, "could not get the bucket")
	}

	var result []refs.Blob

	if err := bucket.ForEach(func(item *badger.Item) error {
		k, err := bucket.KeyInBucket(item)
		if err != nil {
			return errors.Wrap(err, "error determining item key in bucket")
		}

		if k.Len() != 1 {
			return errors.New("invalid length")
		}

		blobRef, err := refs.NewBlob(string(k.Components()[0].Bytes()))
		if err != nil {
			return errors.Wrap(err, "new blob ref error")
		}

		result = append(result, blobRef)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "foreach error")
	}

	return result, nil
}

func (r BlobRepository) ListMessages(blobRef refs.Blob) ([]refs.Message, error) {
	bucket, err := r.createBucketByBlobMessageRefs(blobRef)
	if err != nil {
		return nil, errors.Wrap(err, "could not get the bucket")
	}

	var result []refs.Message

	if err := bucket.ForEach(func(item *badger.Item) error {
		k, err := bucket.KeyInBucket(item)
		if err != nil {
			return errors.Wrap(err, "error determining item key in bucket")
		}

		if k.Len() != 1 {
			return errors.New("invalid length")
		}

		msgRef, err := refs.NewMessage(string(k.Components()[0].Bytes()))
		if err != nil {
			return errors.Wrap(err, "new blob ref error")
		}

		result = append(result, msgRef)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "foreach error")
	}

	return result, nil
}

func (r BlobRepository) createBucketByMessageBlobRefs(ref refs.Message) (utils.Bucket, error) {
	return utils.NewBucket(r.tx, r.bucketPathByMessageBlobRefs(ref))
}

func (r BlobRepository) createBucketByBlobMessageRefs(ref refs.Blob) (utils.Bucket, error) {
	return utils.NewBucket(r.tx, r.bucketPathByBlobMessageRefs(ref))
}

//	func (r BlobRepository) getBucketByMessageBlobRefs(ref refs.Message) (*bbolt.Bucket, error) {
//		return utils.GetBucket(r.tx, r.bucketPathByMessageBlobRefs(ref))
//	}
//func (r BlobRepository) createBucketByBlob() (utils.Bucket, error) {
//	return utils.NewBucket(r.tx, r.bucketPathByBlob())
//}

func (r BlobRepository) bucketPathByMessageBlobRefs(ref refs.Message) utils.Key {
	return r.bucketPathByMessage().Append(utils.MustNewKeyComponent([]byte(ref.String())))
}

func (r BlobRepository) bucketPathByBlobMessageRefs(ref refs.Blob) utils.Key {
	return r.bucketPathByBlob().Append(utils.MustNewKeyComponent([]byte(ref.String())))
}

func (r BlobRepository) bucketPathByMessage() utils.Key {
	return utils.MustNewKey(
		bucketBlobsKeyComponent,
		bucketBlobsByMessageKeyComponent,
	)
}

func (r BlobRepository) bucketPathByBlob() utils.Key {
	return utils.MustNewKey(
		bucketBlobsKeyComponent,
		bucketBlobsByBlobKeyComponent,
	)
}
