package bolt

import (
	"net"
	"strconv"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"go.etcd.io/bbolt"
)

const (
	pubRepositoryBucketPubs                              = "pubs"
	pubRepositoryBucketPubsByPub                         = "by_pub"
	pubRepositoryBucketPubsByMessage                     = "by_message"
	pubRepositoryBucketPubsByPubAddresses                = "addresses"
	pubRepositoryBucketPubsByPubAddressesSources         = "sources"
	pubRepositoryBucketPubsByPubAddressesSourcesMessages = "messages"
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
	byPubBucket, err := r.createBucketByPubMessages(pub)
	if err != nil {
		return errors.Wrap(err, "could not create the by_pub bucket")
	}

	if err := byPubBucket.Put(r.messageKey(pub.Message()), nil); err != nil {
		return errors.Wrap(err, "by_pub bucket put failed")
	}

	pubsBucket, err := r.createBucketByMessagePubs(pub.Message())
	if err != nil {
		return errors.Wrap(err, "could not create the by_message bucket")
	}

	if err := pubsBucket.Put(r.identityKey(pub.Content().Key()), nil); err != nil {
		return errors.Wrap(err, "by_message bucket put failed")
	}

	return nil
}

func (r PubRepository) Delete(msgRef refs.Message) error {
	pubsBucket, err := r.getBucketByMessagePubs(msgRef)
	if err != nil {
		return errors.Wrap(err, "could not get blob refs bucket")
	}
	if pubsBucket == nil {
		return nil
	}

	byPubBucket, err := utils.GetBucket(r.tx, []utils.BucketName{
		utils.BucketName(pubRepositoryBucketPubs),
		utils.BucketName(pubRepositoryBucketPubsByPub),
	})
	if err != nil {
		return errors.Wrap(err, "could not get the by_pub bucket")
	}
	if byPubBucket == nil {
		return nil
	}

	c := pubsBucket.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		identityBucket := byPubBucket.Bucket(k)
		if identityBucket == nil {
			return errors.New("data inconsistency")
		}

		if err := r.removeFromAddresses(identityBucket, msgRef); err != nil {
			return errors.Wrap(err, "failed to remove from addresses")
		}

		if utils.BucketIsEmpty(identityBucket) {
			if err := byPubBucket.DeleteBucket(k); err != nil {
				return errors.Wrap(err, "error deleting the identity bucket")
			}
		}
	}

	if err := utils.DeleteBucket(r.tx, r.bucketPathByMessage(), utils.BucketName(msgRef.String())); err != nil {
		return errors.Wrap(err, "failed to remove from the by_message bucket")
	}

	return nil
}

func (r PubRepository) removeFromAddresses(parentBucket *bbolt.Bucket, msgRef refs.Message) error {
	addressesBucket := parentBucket.Bucket(utils.BucketName(pubRepositoryBucketPubsByPubAddresses))
	if addressesBucket == nil {
		return nil
	}

	c := addressesBucket.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		addressBucket := addressesBucket.Bucket(k)

		if err := r.removeFromSources(addressBucket, msgRef); err != nil {
			return errors.Wrap(err, "failed to remove from sources")
		}

		if utils.BucketIsEmpty(addressBucket) {
			if err := addressesBucket.DeleteBucket(k); err != nil {
				return errors.Wrap(err, "error deleting the address bucket")
			}
		}
	}

	if utils.BucketIsEmpty(addressesBucket) {
		if err := parentBucket.DeleteBucket(utils.BucketName(pubRepositoryBucketPubsByPubAddresses)); err != nil {
			return errors.Wrap(err, "error deleting the sources bucket")
		}
	}

	return nil
}

func (r PubRepository) removeFromSources(parentBucket *bbolt.Bucket, msgRef refs.Message) error {
	sourcesBucket := parentBucket.Bucket(utils.BucketName(pubRepositoryBucketPubsByPubAddressesSources))
	if sourcesBucket == nil {
		return nil
	}

	c := sourcesBucket.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		sourceBucket := sourcesBucket.Bucket(k)

		if err := r.removeFromMessages(sourceBucket, msgRef); err != nil {
			return errors.Wrap(err, "failed to remove from messages")
		}

		if utils.BucketIsEmpty(sourceBucket) {
			if err := sourcesBucket.DeleteBucket(k); err != nil {
				return errors.Wrap(err, "error deleting the message bucket")
			}
		}
	}

	if utils.BucketIsEmpty(sourcesBucket) {
		if err := parentBucket.DeleteBucket(utils.BucketName(pubRepositoryBucketPubsByPubAddressesSources)); err != nil {
			return errors.Wrap(err, "error deleting the sources bucket")
		}
	}

	return nil
}

func (r PubRepository) removeFromMessages(parentBucket *bbolt.Bucket, msgRef refs.Message) error {
	messagesBucket := parentBucket.Bucket(utils.BucketName(pubRepositoryBucketPubsByPubAddressesSourcesMessages))
	if messagesBucket == nil {
		return nil
	}

	if err := messagesBucket.Delete(r.messageKey(msgRef)); err != nil {
		return errors.Wrap(err, "error deleting from the message bucket")
	}

	if utils.BucketIsEmpty(messagesBucket) {
		if err := parentBucket.DeleteBucket(utils.BucketName(pubRepositoryBucketPubsByPubAddressesSourcesMessages)); err != nil {
			return errors.Wrap(err, "error deleting the message bucket")
		}
	}

	return nil
}

func (r PubRepository) messageKey(ref refs.Message) []byte {
	return []byte(ref.String())
}

func (r PubRepository) identityKey(ref refs.Identity) []byte {
	return []byte(ref.String())
}

func (r PubRepository) addressAsString(pub content.Pub) string {
	return net.JoinHostPort(pub.Host(), strconv.Itoa(pub.Port()))
}

func (r PubRepository) createBucketByPubMessages(pub feeds.PubToSave) (*bbolt.Bucket, error) {
	return utils.CreateBucket(r.tx, r.bucketPathByPubMessages(pub.Who(), pub.Content()))
}

func (r PubRepository) getBucketByPubMessages(pub feeds.PubToSave) (*bbolt.Bucket, error) {
	return utils.GetBucket(r.tx, r.bucketPathByPubMessages(pub.Who(), pub.Content()))
}

func (r PubRepository) createBucketByMessagePubs(ref refs.Message) (*bbolt.Bucket, error) {
	return utils.CreateBucket(r.tx, r.bucketPathByMessagePubs(ref))
}

func (r PubRepository) getBucketByMessagePubs(ref refs.Message) (*bbolt.Bucket, error) {
	return utils.GetBucket(r.tx, r.bucketPathByMessagePubs(ref))
}

func (r PubRepository) bucketPathByPubMessages(source refs.Identity, pub content.Pub) []utils.BucketName {
	return []utils.BucketName{
		utils.BucketName(pubRepositoryBucketPubs),
		utils.BucketName(pubRepositoryBucketPubsByPub),
		utils.BucketName(pub.Key().String()),
		utils.BucketName(pubRepositoryBucketPubsByPubAddresses),
		utils.BucketName(r.addressAsString(pub)),
		utils.BucketName(pubRepositoryBucketPubsByPubAddressesSources),
		utils.BucketName(source.String()),
		utils.BucketName(pubRepositoryBucketPubsByPubAddressesSourcesMessages),
	}
}

func (r PubRepository) bucketPathByMessage() []utils.BucketName {
	return []utils.BucketName{
		utils.BucketName(pubRepositoryBucketPubs),
		utils.BucketName(pubRepositoryBucketPubsByMessage),
	}
}

func (r PubRepository) bucketPathByMessagePubs(ref refs.Message) []utils.BucketName {
	return []utils.BucketName{
		utils.BucketName(pubRepositoryBucketPubs),
		utils.BucketName(pubRepositoryBucketPubsByMessage),
		utils.BucketName(ref.String()),
	}
}
