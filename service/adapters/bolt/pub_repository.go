package bolt

import (
	"net"
	"strconv"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/feeds"
	"github.com/planetary-social/go-ssb/service/domain/feeds/content"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"go.etcd.io/bbolt"
)

type PubRepository struct {
	tx *bbolt.Tx
}

func NewPubRepository(
	tx *bbolt.Tx,
) *PubRepository {
	return &PubRepository{
		tx: tx,
	}
}

func (r PubRepository) Put(pub feeds.PubToSave) error {
	messageBucket, err := r.createMessageBucket(pub)
	if err != nil {
		return errors.Wrap(err, "could not create the bucket")
	}

	key := r.messageKey(pub.Id())

	if err := messageBucket.Put(key, nil); err != nil {
		return errors.Wrap(err, "bucket put failed")
	}

	return nil
}

func (r PubRepository) messageKey(id refs.Message) []byte {
	return []byte(id.String())
}

func (r PubRepository) createMessageBucket(pub feeds.PubToSave) (*bbolt.Bucket, error) {
	return createBucket(r.tx, r.messageBucketPath(pub.Who(), pub.Content()))
}

func (r PubRepository) messageBucketPath(source refs.Identity, pub content.Pub) []bucketName {
	return []bucketName{
		bucketName("pubs"),
		bucketName(pub.Key().String()),
		bucketName("addresses"),
		bucketName(net.JoinHostPort(pub.Host(), strconv.Itoa(pub.Port()))),
		bucketName("sources"),
		bucketName(source.String()),
		bucketName("messages"),
	}
}
