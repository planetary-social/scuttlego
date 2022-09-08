package bolt

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"go.etcd.io/bbolt"
)

var (
	bucketBlobs          = utils.BucketName("blobs")
	bucketBlobsByMessage = utils.BucketName("by_message")
	bucketBlobsByBlob    = utils.BucketName("by_blob")
)

type BlobRepository struct {
	tx *bbolt.Tx
}

func NewBlobRepository(tx *bbolt.Tx) *BlobRepository {
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
		if err := byMessageBucket.Put([]byte(blobRef.String()), nil); err != nil {
			return errors.Wrap(err, "by_message bucket put failed")
		}

		byBlobBucket, err := r.createBucketByBlobMessageRefs(blobRef)
		if err != nil {
			return errors.Wrap(err, "could not create by_blob bucket")
		}

		if err := byBlobBucket.Put([]byte(msgRef.String()), nil); err != nil {
			return errors.Wrap(err, "by_blob bucket put failed")
		}
	}

	return nil
}

func (r BlobRepository) Delete(msgRef refs.Message) error {
	byMessageBlobRefsBucket, err := r.getBucketByMessageBlobRefs(msgRef)
	if err != nil {
		return errors.Wrap(err, "could not get blob refs bucket")
	}

	if byMessageBlobRefsBucket == nil {
		return nil
	}

	byBlobBucket, err := r.getBucketByBlob()
	if err != nil {
		return errors.Wrap(err, "could not get the by blob bucket")
	}

	if byBlobBucket == nil {
		return errors.New("invalid state, one bucket exists but the other one does not")
	}

	c := byMessageBlobRefsBucket.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		blobRefBucket := byBlobBucket.Bucket(k)

		if err := blobRefBucket.Delete([]byte(msgRef.String())); err != nil {
			return errors.Wrap(err, "could not remove from the bucket")
		}

		if utils.BucketIsEmpty(blobRefBucket) {
			if err := byBlobBucket.DeleteBucket(k); err != nil {
				return errors.Wrap(err, "could not remove the blob ref bucket")
			}
		}
	}

	if err := utils.DeleteBucket(r.tx, r.bucketPathByMessage(), utils.BucketName(msgRef.String())); err != nil {
		return errors.Wrap(err, "could not delete the message bucket")
	}

	return nil
}

func (r BlobRepository) createBucketByMessageBlobRefs(ref refs.Message) (*bbolt.Bucket, error) {
	return utils.CreateBucket(r.tx, r.bucketPathByMessageBlobRefs(ref))
}

func (r BlobRepository) createBucketByBlobMessageRefs(ref refs.Blob) (*bbolt.Bucket, error) {
	return utils.CreateBucket(r.tx, r.bucketPathByBlobMessageRefs(ref))
}

func (r BlobRepository) getBucketByMessageBlobRefs(ref refs.Message) (*bbolt.Bucket, error) {
	return utils.GetBucket(r.tx, r.bucketPathByMessageBlobRefs(ref))
}

func (r BlobRepository) getBucketByBlobMessageRefs(ref refs.Blob) (*bbolt.Bucket, error) {
	return utils.GetBucket(r.tx, r.bucketPathByBlobMessageRefs(ref))
}

func (r BlobRepository) getBucketByBlob() (*bbolt.Bucket, error) {
	return utils.GetBucket(r.tx, r.bucketPathByBlob())
}

func (r BlobRepository) bucketPathByMessageBlobRefs(ref refs.Message) []utils.BucketName {
	return []utils.BucketName{
		bucketBlobs,
		bucketBlobsByMessage,
		utils.BucketName(ref.String()),
	}
}

func (r BlobRepository) bucketPathByBlobMessageRefs(ref refs.Blob) []utils.BucketName {
	return []utils.BucketName{
		bucketBlobs,
		bucketBlobsByBlob,
		utils.BucketName(ref.String()),
	}
}

func (r BlobRepository) bucketPathByMessage() []utils.BucketName {
	return []utils.BucketName{
		bucketBlobs,
		bucketBlobsByMessage,
	}
}

func (r BlobRepository) bucketPathByBlob() []utils.BucketName {
	return []utils.BucketName{
		bucketBlobs,
		bucketBlobsByBlob,
	}
}
