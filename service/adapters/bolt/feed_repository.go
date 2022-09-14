package bolt

import (
	"encoding/binary"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"go.etcd.io/bbolt"
)

var ErrFeedNotFound = errors.New("feed not found")

type FeedRepository struct {
	tx                *bbolt.Tx
	graph             *SocialGraphRepository
	receiveLog        *ReceiveLogRepository
	messageRepository *MessageRepository
	pubRepository     *PubRepository
	blobRepository    *BlobRepository
	banListRepository *BanListRepository
	formatScuttlebutt *formats.Scuttlebutt
}

func NewFeedRepository(
	tx *bbolt.Tx,
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

	return b.saveFeed(ref, feed)
}

func (b FeedRepository) GetFeed(ref refs.Feed) (*feeds.Feed, error) {
	f, err := b.loadFeed(ref)
	if err != nil {
		return nil, errors.Wrap(err, "loading failed")
	}

	if f == nil {
		return nil, ErrFeedNotFound
	}

	return f, nil
}

func (b FeedRepository) GetMessages(id refs.Feed, seq *message.Sequence, limit *int) ([]message.Message, error) {
	var messages []message.Message

	bucket, err := b.getFeedBucket(id)
	if err != nil {
		return nil, errors.Wrap(err, "could not get the bucket")
	}

	if bucket == nil {
		return nil, nil
	}

	// todo not stupid implementation (with a cursor)

	if err := bucket.ForEach(func(key, value []byte) error {
		msgId, err := refs.NewMessage(string(value))
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
	bucket, err := b.getFeedsBucket()
	if err != nil {
		return 0, errors.Wrap(err, "could not get the bucket")
	}

	if bucket == nil {
		return 0, nil
	}

	var result int

	// todo probably a read model
	if err := bucket.ForEach(func(k, v []byte) error {
		result++
		return nil
	}); err != nil {
		return 0, errors.Wrap(err, "iteration failed")
	}

	return result, nil
}

func (b FeedRepository) DeleteFeed(ref refs.Feed) error {
	bucket, err := b.getFeedBucket(ref)
	if err != nil {
		return errors.Wrap(err, "could not get the bucket")
	}

	if bucket == nil {
		return nil
	}

	c := bucket.Cursor()

	for k, value := c.First(); k != nil; k, value = c.Next() {
		msgId, err := refs.NewMessage(string(value))
		if err != nil {
			return errors.Wrap(err, "failed to create a message ref")
		}

		if err := b.removeMessageData(msgId); err != nil {
			return errors.Wrap(err, "failed to remove message data")
		}
	}

	if err := b.removeFeedData(ref); err != nil {
		return errors.Wrap(err, "failed to remove feed data")
	}

	return nil
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

	if err := b.deleteFeedBucket(ref); err != nil {
		return errors.Wrap(err, "failed to remove from feed repository")
	}

	return nil
}

func (b FeedRepository) loadFeed(ref refs.Feed) (*feeds.Feed, error) {
	bucket, err := b.getFeedBucket(ref)
	if err != nil {
		return nil, errors.Wrap(err, "could not get the bucket")
	}

	if bucket == nil {
		return nil, nil
	}

	key, value := bucket.Cursor().Last()

	if key == nil && value == nil {
		return nil, nil // to be honest this should not be possible anyway as buckets are created only when saving
	}

	msgId, err := refs.NewMessage(string(value))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a message ref")
	}

	msg, err := b.messageRepository.Get(msgId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the message")
	}

	feed, err := feeds.NewFeedFromHistory(msg, b.formatScuttlebutt)
	if err != nil {
		return nil, errors.Wrap(err, "could not recreate a feed from history")
	}

	return feed, nil
}

func (b FeedRepository) saveFeed(ref refs.Feed, feed *feeds.Feed) error {
	msgsToPersist := feed.PopForPersisting()

	if len(msgsToPersist) != 0 {
		bucket, err := b.createFeedBucket(ref)
		if err != nil {
			return errors.Wrap(err, "could not create the bucket")
		}

		if err := b.banListRepository.CreateFeedMapping(ref); err != nil {
			return errors.Wrap(err, "failed to create the ban list mapping")
		}

		for _, msgToPersist := range msgsToPersist {
			if err := b.saveMessageInBucket(bucket, msgToPersist.Message()); err != nil {
				return errors.Wrap(err, "could not save a message in bucket")
			}

			if err := b.saveMessageInRepositories(msgToPersist); err != nil {
				return errors.Wrap(err, "could not save a message in repositories")
			}
		}
	}

	return nil
}

func (b FeedRepository) saveMessageInRepositories(msg feeds.MessageToPersist) error {
	if err := b.messageRepository.Put(msg.Message()); err != nil {
		return errors.Wrap(err, "message repository put failed")
	}

	if err := b.receiveLog.Put(msg.Message().Id()); err != nil {
		return errors.Wrap(err, "receive log put failed")
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

func (b FeedRepository) saveMessageInBucket(bucket *bbolt.Bucket, msg message.Message) error {
	key := messageKey(msg.Sequence())
	value := []byte(msg.Id().String())

	if err := bucket.Put(key, value); err != nil {
		return errors.Wrap(err, "writing to the bucket failed")
	}

	return nil
}

func (b FeedRepository) saveContact(contact feeds.ContactToSave) error {
	return b.graph.UpdateContact(contact.Who(), contact.Msg().Contact(), func(c *feeds.Contact) error {
		return c.Update(contact.Msg().Actions())
	})
}

func messageKey(seq message.Sequence) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(seq.Int()))
	return buf
}

func (b FeedRepository) createFeedBucket(ref refs.Feed) (bucket *bbolt.Bucket, err error) {
	return utils.CreateBucket(b.tx, b.feedBucketPath(ref))
}

func (b FeedRepository) getFeedBucket(ref refs.Feed) (*bbolt.Bucket, error) {
	return utils.GetBucket(b.tx, b.feedBucketPath(ref))
}

func (b FeedRepository) deleteFeedBucket(ref refs.Feed) error {
	return utils.DeleteBucket(b.tx, b.feedsBucketPath(), utils.BucketName(ref.String()))
}

func (b FeedRepository) getFeedsBucket() (*bbolt.Bucket, error) {
	return utils.GetBucket(b.tx, b.feedsBucketPath())
}

func (b FeedRepository) feedsBucketPath() []utils.BucketName {
	return []utils.BucketName{
		utils.BucketName("feeds"),
	}
}

func (b FeedRepository) feedBucketPath(ref refs.Feed) []utils.BucketName {
	return []utils.BucketName{
		utils.BucketName("feeds"),
		utils.BucketName(ref.String()),
	}
}
