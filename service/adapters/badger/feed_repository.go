package badger

import (
	"encoding/binary"
	"fmt"

	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/planetary-social/scuttlego/service/adapters/badger/utils"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type FeedRepository struct {
	tx                *badger.Txn
	graph             *SocialGraphRepository
	receiveLog        *ReceiveLogRepository
	messageRepository *MessageRepository
	pubRepository     *PubRepository
	blobRepository    *BlobRepository
	banListRepository *BanListRepository
	formatScuttlebutt *formats.Scuttlebutt
}

func NewFeedRepository(
	tx *badger.Txn,
	graph *SocialGraphRepository,
	receiveLog *ReceiveLogRepository,
	messageRepository *MessageRepository,
	pubRepository *PubRepository,
	blobRepository *BlobRepository,
	banListRepository *BanListRepository,
	formatScuttlebutt *formats.Scuttlebutt,
) *FeedRepository {
	return &FeedRepository{
		tx:                tx,
		graph:             graph,
		receiveLog:        receiveLog,
		messageRepository: messageRepository,
		pubRepository:     pubRepository,
		blobRepository:    blobRepository,
		banListRepository: banListRepository,
		formatScuttlebutt: formatScuttlebutt,
	}
}

func (b FeedRepository) UpdateFeed(ref refs.Feed, f commands.UpdateFeedFn) error {
	feed, err := b.loadFeed(ref)
	if err != nil {
		return errors.Wrap(err, "could not load a feed")
	}

	if feed == nil {
		feed = feeds.NewFeed(b.formatScuttlebutt)
	}

	if err = f(feed); err != nil {
		return errors.Wrap(err, "provided function returned an error")
	}

	return b.saveFeed(ref, feed, true)
}

func (b FeedRepository) UpdateFeedIgnoringReceiveLog(ref refs.Feed, f commands.UpdateFeedFn) error {
	feed, err := b.loadFeed(ref)
	if err != nil {
		return errors.Wrap(err, "could not load a feed")
	}

	if feed == nil {
		feed = feeds.NewFeed(b.formatScuttlebutt)
	}

	if err = f(feed); err != nil {
		return errors.Wrap(err, "provided function returned an error")
	}

	return b.saveFeed(ref, feed, false)
}

func (b FeedRepository) GetFeed(ref refs.Feed) (*feeds.Feed, error) {
	f, err := b.loadFeed(ref)
	if err != nil {
		return nil, errors.Wrap(err, "loading failed")
	}

	if f == nil {
		return nil, common.ErrFeedNotFound
	}

	return f, nil
}

func (b FeedRepository) GetMessages(id refs.Feed, seq *message.Sequence, limit *int) ([]message.Message, error) {
	var messages []message.Message

	bucket := b.getFeedBucket(id)

	// todo not stupid implementation (with seek)
	if err := bucket.ForEach(func(item *badger.Item) error {
		valueCopy, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Wrap(err, "error getting value")

		}

		msgId, err := refs.NewMessage(string(valueCopy))
		if err != nil {
			return errors.Wrap(err, "failed to create a message ref")
		}

		msg, err := b.messageRepository.Get(msgId)
		if err != nil {
			return errors.Wrap(err, "failed to get the message")
		}

		if (limit == nil || len(messages) < *limit) && (seq == nil || !seq.ComesAfter(msg.Sequence())) {
			messages = append(messages, msg)
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to iterate")
	}

	return messages, nil
}

func (b FeedRepository) Count() (int, error) {
	c, err := b.getFeedCounter()
	if err != nil {
		return 0, errors.Wrap(err, "could not get the counter")
	}

	v, err := c.Get()
	if err != nil {
		return 0, errors.Wrap(err, "could not get the count")
	}

	return int(v), nil
}

func (b FeedRepository) DeleteFeed(ref refs.Feed) error {
	bucket := b.getFeedBucket(ref)

	if err := bucket.ForEach(func(item *badger.Item) error {
		valueCopy, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Wrap(err, "error getting value")
		}

		msgId, err := refs.NewMessage(string(valueCopy))
		if err != nil {
			return errors.Wrap(err, "failed to create a message ref")
		}

		if err := b.removeMessageData(msgId); err != nil {
			return errors.Wrap(err, "failed to remove message data")
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "foreach error")
	}

	if err := b.removeFeedData(ref); err != nil {
		return errors.Wrap(err, "failed to remove feed data")
	}

	return nil
}

func (b FeedRepository) GetMessage(feed refs.Feed, sequence message.Sequence) (message.Message, error) {
	bucket := b.getFeedBucket(feed)

	item, err := bucket.Get(b.marshalMessageKey(sequence))
	if err != nil {
		return message.Message{}, errors.Wrap(err, "sequence not found")
	}

	var msgId refs.Message

	if err := item.Value(func(val []byte) error {
		tmp, err := refs.NewMessage(string(val))
		if err != nil {
			return errors.Wrap(err, "failed to create a message ref")
		}
		msgId = tmp
		return nil
	}); err != nil {
		return message.Message{}, errors.Wrap(err, "error getting value")
	}

	return b.messageRepository.Get(msgId)
}

func (b FeedRepository) removeMessageData(ref refs.Message) error {
	if err := b.messageRepository.Delete(ref); err != nil {
		return errors.Wrap(err, "failed to remove from message repository")
	}

	if err := b.blobRepository.Delete(ref); err != nil {
		return errors.Wrap(err, "failed to remove from blob repository")
	}

	if err := b.pubRepository.Delete(ref); err != nil {
		return errors.Wrap(err, "failed to remove from pub repository")
	}

	return nil
}

func (b FeedRepository) removeFeedData(ref refs.Feed) error {
	idenRef, err := refs.NewIdentityFromPublic(ref.Identity()) // todo figure out if this should be feed or identity
	if err != nil {
		return errors.Wrap(err, "failed to create an identity ref")
	}

	if err := b.graph.Remove(idenRef); err != nil {
		return errors.Wrap(err, "failed to remove from graph repository")
	}

	if err := b.banListRepository.RemoveFeedMapping(ref); err != nil {
		return errors.Wrap(err, "failed to remove the ban list mapping")
	}

	if !b.getFeedBucket(ref).IsEmpty() {
		c, err := b.getFeedCounter()
		if err != nil {
			return errors.Wrap(err, "could not get the counter")
		}

		if err := c.Decrement(); err != nil {
			return errors.Wrap(err, "error decrementing the feed counter")
		}
	}

	if err := b.getFeedBucket(ref).DeleteBucket(); err != nil {
		return errors.Wrap(err, "failed to remove from feed bucket")
	}

	return nil
}

func (b FeedRepository) loadFeed(ref refs.Feed) (*feeds.Feed, error) {
	bucket := b.getFeedBucket(ref)

	it := bucket.IteratorWithModifiedOptions(func(options *badger.IteratorOptions) {
		options.Reverse = true
		options.PrefetchValues = false
	})
	defer it.Close()

	it.Seek(bucket.Prefix().Bytes())
	if !it.Valid() {
		return nil, nil
	}

	valueCopy, err := it.Item().ValueCopy(nil)
	if err != nil {
		return nil, errors.Wrap(err, "error getting value")
	}

	msgId, err := refs.NewMessage(string(valueCopy))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a message ref")
	}

	msg, err := b.messageRepository.Get(msgId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get message '%s'", msgId)
	}

	feed, err := feeds.NewFeedFromHistory(msg, b.formatScuttlebutt)
	if err != nil {
		return nil, errors.Wrap(err, "could not recreate a feed from history")
	}

	return feed, nil
}

type FeedMessage struct {
	Sequence message.Sequence
	Id       refs.Message
}

func (b FeedRepository) GetFeedMessages(ref refs.Feed) ([]FeedMessage, error) {
	bucket := b.getFeedBucket(ref)

	var result []FeedMessage

	if err := bucket.ForEach(func(item *badger.Item) error {
		keyInBucket, err := bucket.KeyInBucket(item)
		if err != nil {
			return errors.Wrap(err, "error getting key in bucket")
		}

		sequence, err := b.unmarshalMessageKey(keyInBucket.Bytes())
		if err != nil {
			return errors.Wrap(err, "error unmarshaling sequence")
		}

		valueCopy, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Wrap(err, "error getting value copy")
		}

		id, err := refs.NewMessage(string(valueCopy))
		if err != nil {
			return errors.Wrap(err, "error creating a message ref")
		}

		result = append(result, FeedMessage{
			Sequence: sequence,
			Id:       id,
		})

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "foreach error")
	}

	return result, nil
}

func (b FeedRepository) saveFeed(ref refs.Feed, feed *feeds.Feed, saveInReceiveLog bool) error {
	msgsToPersist := feed.PopForPersisting()

	if len(msgsToPersist) != 0 {
		bucket := b.getFeedBucket(ref)

		if bucket.IsEmpty() {
			c, err := b.getFeedCounter()
			if err != nil {
				return errors.Wrap(err, "could not get the counter")
			}

			if err := c.Increment(); err != nil {
				return errors.Wrap(err, "error incrementing the feed counter")
			}
		}

		if err := b.banListRepository.CreateFeedMapping(ref); err != nil {
			return errors.Wrap(err, "failed to create the ban list mapping")
		}

		for _, msgToPersist := range msgsToPersist {
			if err := b.saveMessageInBucket(bucket, msgToPersist.Message()); err != nil {
				return errors.Wrap(err, "could not save a message in bucket")
			}

			if err := b.saveMessageInRepositories(msgToPersist, saveInReceiveLog); err != nil {
				return errors.Wrap(err, "could not save a message in repositories")
			}
		}
	}

	return nil
}

func (b FeedRepository) saveMessageInRepositories(msg feeds.MessageToPersist, saveInReceiveLog bool) error {
	if err := b.messageRepository.Put(msg.Message()); err != nil {
		return errors.Wrap(err, "message repository put failed")
	}

	if saveInReceiveLog {
		if err := b.receiveLog.Put(msg.Message().Id()); err != nil {
			return errors.Wrap(err, "receive log put failed")
		}
	}

	for _, contact := range msg.ContactsToSave() {
		if err := b.saveContact(contact); err != nil {
			return errors.Wrap(err, "failed to save a contact")
		}
	}

	for _, pub := range msg.PubsToSave() {
		if err := b.pubRepository.Put(pub); err != nil {
			return errors.Wrap(err, "pub repository put failed")
		}
	}

	for _, blob := range msg.BlobsToSave() {
		if err := b.blobRepository.Put(msg.Message().Id(), blob); err != nil {
			return errors.Wrap(err, "blob repository put failed")
		}
	}

	return nil
}

func (b FeedRepository) saveMessageInBucket(bucket utils.Bucket, msg message.Message) error {
	key := b.marshalMessageKey(msg.Sequence())
	value := []byte(msg.Id().String())

	if err := bucket.Set(key, value); err != nil {
		return errors.Wrap(err, "writing to the bucket failed")
	}

	return nil
}

func (b FeedRepository) saveContact(contact feeds.ContactToSave) error {
	return b.graph.UpdateContact(contact.Who(), contact.Msg().Contact(), func(c *feeds.Contact) error {
		return c.Update(contact.Msg().Actions())
	})
}

const messageKeyLen = 8

func (b FeedRepository) marshalMessageKey(seq message.Sequence) []byte {
	buf := make([]byte, messageKeyLen)
	binary.BigEndian.PutUint64(buf, uint64(seq.Int()))
	return buf
}

func (b FeedRepository) unmarshalMessageKey(marshaledKey []byte) (message.Sequence, error) {
	if l := len(marshaledKey); l != messageKeyLen {
		return message.Sequence{}, fmt.Errorf("invalid message key length '%d'", l)
	}
	v := binary.BigEndian.Uint64(marshaledKey)
	return message.NewSequence(int(v))
}

func (b FeedRepository) getFeedCounter() (utils.Counter, error) {
	return utils.NewCounter(b.getMetaBucket(), utils.MustNewKeyComponent([]byte("feed_count")))
}

func (b FeedRepository) getFeedBucket(ref refs.Feed) utils.Bucket {
	return utils.MustNewBucket(b.tx, b.feedBucketPath(ref))
}

func (b FeedRepository) getMetaBucket() utils.Bucket {
	return utils.MustNewBucket(b.tx, b.metaBucketPath())
}

func (b FeedRepository) feedBucketPath(ref refs.Feed) utils.Key {
	return utils.MustNewKey(
		utils.MustNewKeyComponent([]byte("feeds")),
		utils.MustNewKeyComponent([]byte("entries")),
		utils.MustNewKeyComponent([]byte(ref.String())),
	)
}

func (b FeedRepository) metaBucketPath() utils.Key {
	return utils.MustNewKey(
		utils.MustNewKeyComponent([]byte("feeds")),
		utils.MustNewKeyComponent([]byte("meta")),
	)
}
