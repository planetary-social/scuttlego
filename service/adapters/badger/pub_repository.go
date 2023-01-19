package badger

import (
	"net"
	"strconv"

	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/planetary-social/scuttlego/service/adapters/badger/utils"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

const (
	pubRepositoryBucketPubs = "pubs"

	pubRepositoryBucketPubsByPub                         = "by_pub"
	pubRepositoryBucketPubsByPubAddresses                = "addresses"
	pubRepositoryBucketPubsByPubAddressesSources         = "sources"
	pubRepositoryBucketPubsByPubAddressesSourcesMessages = "messages"

	pubRepositoryBucketPubsByMessage = "by_message"
)

type PubRepository struct {
	tx *badger.Txn
}

func NewPubRepository(
	tx *badger.Txn,
) *PubRepository {
	return &PubRepository{
		tx: tx,
	}
}

func (r PubRepository) Put(pub feeds.PubToSave) error {
	byPubBucket := r.createBucketByPubMessages(pub)
	if err := byPubBucket.Set(r.messageKey(pub.Message()), nil); err != nil {
		return errors.Wrap(err, "by_pub bucket put failed")
	}

	pubsBucket := r.createBucketByMessagePubs(pub.Message())
	if err := pubsBucket.Set(r.identityKey(pub.Content().Key()), nil); err != nil {
		return errors.Wrap(err, "by_message bucket put failed")
	}

	return nil
}

func (r PubRepository) Delete(msgRef refs.Message) error {
	pubsBucket := r.createBucketByMessagePubs(msgRef)

	if err := pubsBucket.ForEach(func(item *badger.Item) error {
		pubIdentityKey, err := pubsBucket.KeyInBucket(item)
		if err != nil {
			return errors.Wrap(err, "error determining key in bucket")
		}

		pubIdentityRef, err := refs.NewIdentity(string(pubIdentityKey.Bytes()))
		if err != nil {
			return errors.Wrap(err, "error creating pub identity ref")
		}

		if err := r.removeFromByPub(pubIdentityRef, msgRef); err != nil {
			return errors.Wrap(err, "unable to remove from by pub")
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "foreach error")
	}

	if err := pubsBucket.DeleteBucket(); err != nil {
		return errors.Wrap(err, "failed to remove from the by_message bucket")
	}

	return nil
}

func (r PubRepository) ListPubs(msgRef refs.Message) ([]refs.Identity, error) {
	pubsBucket := r.createBucketByMessagePubs(msgRef)

	var result []refs.Identity

	if err := pubsBucket.ForEach(func(item *badger.Item) error {
		pubIdentityKey, err := pubsBucket.KeyInBucket(item)
		if err != nil {
			return errors.Wrap(err, "error determining key in bucket")
		}

		pubIdentityRef, err := refs.NewIdentity(string(pubIdentityKey.Bytes()))
		if err != nil {
			return errors.Wrap(err, "error creating pub identity ref")
		}

		result = append(result, pubIdentityRef)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "foreach error")
	}

	return result, nil
}

type PubAddress struct {
	Address string
	Sources []PubAddressSource
}

type PubAddressSource struct {
	Identity refs.Identity
	Messages []refs.Message
}

func (r PubRepository) ListAddresses(pubIdentityRef refs.Identity) ([]PubAddress, error) {
	byPubIdentityBucket := utils.MustNewBucket(r.tx, r.bucketPathByPub(pubIdentityRef))

	var result []PubAddress

	if err := byPubIdentityBucket.ForEach(func(item *badger.Item) error {
		key, err := utils.NewKeyFromBytes(item.KeyCopy(nil))
		if err != nil {
			return errors.Wrap(err, "error creating a key")
		}

		if key.Len() != 9 {
			return errors.New("invalid key length")
		}

		addressComponent := key.Components()[key.Len()-5]
		sourceIdentityRefComponent := key.Components()[key.Len()-3]
		messageRefComponent := key.Components()[key.Len()-1]

		itemAddress := string(addressComponent.Bytes())

		itemSourceIdentityRef, err := refs.NewIdentity(string(sourceIdentityRefComponent.Bytes()))
		if err != nil {
			return errors.Wrap(err, "error creating message ref")
		}

		itemMessageRef, err := refs.NewMessage(string(messageRefComponent.Bytes()))
		if err != nil {
			return errors.Wrap(err, "error creating message ref")
		}

		result = r.mergeListAddressesResults(result, itemAddress, itemSourceIdentityRef, itemMessageRef)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "foreach error")
	}

	return result, nil
}

func (r PubRepository) mergeListAddressesResults(result []PubAddress, address string, source refs.Identity, msg refs.Message) []PubAddress {
	for i := range result {
		if result[i].Address == address {
			for j := range result[i].Sources {
				if result[i].Sources[j].Identity.Equal(source) {
					result[i].Sources[j].Messages = append(result[i].Sources[j].Messages, msg)
					return result
				}
			}

			result[i].Sources = append(result[i].Sources, PubAddressSource{
				Identity: source,
				Messages: []refs.Message{msg},
			})
			return result
		}
	}

	return append(result, PubAddress{
		Address: address,
		Sources: []PubAddressSource{
			{
				Identity: source,
				Messages: []refs.Message{msg},
			},
		},
	})
}

func (r PubRepository) removeFromByPub(pubIdentityRef refs.Identity, msgRef refs.Message) error {
	byPubIdentityBucket := utils.MustNewBucket(r.tx, r.bucketPathByPub(pubIdentityRef))

	if err := byPubIdentityBucket.ForEach(func(item *badger.Item) error {
		key, err := utils.NewKeyFromBytes(item.KeyCopy(nil))
		if err != nil {
			return errors.Wrap(err, "error creating a key")
		}

		lastComponent := key.Components()[key.Len()-1]

		itemMessageRef, err := refs.NewMessage(string(lastComponent.Bytes()))
		if err != nil {
			return errors.Wrap(err, "error creating message ref")
		}

		if itemMessageRef.Equal(msgRef) {
			if err := r.tx.Delete(item.KeyCopy(nil)); err != nil {
				return errors.Wrap(err, "delete error")
			}
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "foreach error")
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

func (r PubRepository) createBucketByPubMessages(pub feeds.PubToSave) utils.Bucket {
	return utils.MustNewBucket(r.tx, r.bucketPathByPubMessages(pub.Who(), pub.Content()))
}

func (r PubRepository) createBucketByMessagePubs(ref refs.Message) utils.Bucket {
	return utils.MustNewBucket(r.tx, r.bucketPathByMessagePubs(ref))
}

func (r PubRepository) bucketPathByPub(pubIdentity refs.Identity) utils.Key {
	return utils.MustNewKey(
		utils.MustNewKeyComponent([]byte(pubRepositoryBucketPubs)),
		utils.MustNewKeyComponent([]byte(pubRepositoryBucketPubsByPub)),
		utils.MustNewKeyComponent([]byte(pubIdentity.String())),
	)
}

func (r PubRepository) bucketPathByPubMessages(source refs.Identity, pub content.Pub) utils.Key {
	return utils.MustNewKey(
		utils.MustNewKeyComponent([]byte(pubRepositoryBucketPubs)),
		utils.MustNewKeyComponent([]byte(pubRepositoryBucketPubsByPub)),
		utils.MustNewKeyComponent([]byte(pub.Key().String())),
		utils.MustNewKeyComponent([]byte(pubRepositoryBucketPubsByPubAddresses)),
		utils.MustNewKeyComponent([]byte(r.addressAsString(pub))),
		utils.MustNewKeyComponent([]byte(pubRepositoryBucketPubsByPubAddressesSources)),
		utils.MustNewKeyComponent([]byte(source.String())),
		utils.MustNewKeyComponent([]byte(pubRepositoryBucketPubsByPubAddressesSourcesMessages)),
	)
}

func (r PubRepository) bucketPathByMessagePubs(ref refs.Message) utils.Key {
	return utils.MustNewKey(
		utils.MustNewKeyComponent([]byte(pubRepositoryBucketPubs)),
		utils.MustNewKeyComponent([]byte(pubRepositoryBucketPubsByMessage)),
		utils.MustNewKeyComponent([]byte(ref.String())),
	)
}
