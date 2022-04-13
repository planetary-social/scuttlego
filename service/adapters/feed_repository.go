package adapters

import (
	"encoding/binary"
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/feeds"
	"github.com/planetary-social/go-ssb/service/domain/feeds/content"
	"github.com/planetary-social/go-ssb/service/domain/feeds/formats"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/replication"
	"go.etcd.io/bbolt"
)

type BoltFeedRepository struct {
	tx                *bbolt.Tx
	graph             *SocialGraphRepository
	receiveLog        *ReceiveLogRepository
	messageRepository *BoltMessageRepository
	formatScuttlebutt *formats.Scuttlebutt
}

func NewBoltFeedRepository(
	tx *bbolt.Tx,
	identifier RawMessageIdentifier,
	graph *SocialGraphRepository,
	receiveLog *ReceiveLogRepository,
	messageRepository *BoltMessageRepository,
	formatScuttlebutt *formats.Scuttlebutt,
) *BoltFeedRepository {
	return &BoltFeedRepository{
		tx:                tx,
		graph:             graph,
		receiveLog:        receiveLog,
		messageRepository: messageRepository,
		formatScuttlebutt: formatScuttlebutt,
	}
}

func (b BoltFeedRepository) UpdateFeed(ref refs.Feed, f func(feed *feeds.Feed) (*feeds.Feed, error)) error {
	feed, err := b.loadFeed(ref)
	if err != nil {
		return errors.Wrap(err, "could not load a feed")
	}

	if feed == nil {
		feed = feeds.NewFeed(b.formatScuttlebutt)
	}

	feed, err = f(feed)
	if err != nil {
		return errors.Wrap(err, "provided function returned an error")
	}

	return b.saveFeed(ref, feed)
}
func (b BoltFeedRepository) GetFeed(ref refs.Feed) (*feeds.Feed, error) {
	f, err := b.loadFeed(ref)
	if err != nil {
		return nil, errors.Wrap(err, "loading failed")
	}

	if f == nil {
		return nil, replication.ErrFeedNotFound
	}

	return f, nil
}

func (b BoltFeedRepository) loadFeed(ref refs.Feed) (*feeds.Feed, error) {
	bucket, err := getFeedBucket(b.tx, ref)
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

func (b BoltFeedRepository) saveFeed(ref refs.Feed, feed *feeds.Feed) error {
	msgs, contacts := feed.PopForPersisting()

	if len(msgs) != 0 {
		bucket, err := createFeedBucket(b.tx, ref)
		if err != nil {
			return errors.Wrap(err, "could not create the bucket")
		}

		for _, msg := range msgs {
			if err := b.saveMessage(bucket, msg); err != nil {
				return errors.Wrap(err, "could not save a message")
			}
		}
	}

	for _, contact := range contacts {
		if err := b.saveContact(contact); err != nil {
			return errors.Wrap(err, "failed to save a contact")
		}
	}

	return nil
}

func (b BoltFeedRepository) saveContact(contact feeds.ContactToSave) error {
	switch contact.Msg().Action() {
	case content.ContactActionFollow:
		return b.graph.Follow(contact.Who(), contact.Msg().Contact())
	case content.ContactActionUnfollow:
		return b.graph.Unfollow(contact.Who(), contact.Msg().Contact())
	case content.ContactActionBlock:
		return b.graph.Block(contact.Who(), contact.Msg().Contact())
	case content.ContactActionUnblock:
		return b.graph.Unblock(contact.Who(), contact.Msg().Contact())
	default:
		return fmt.Errorf("unknown contact action '%#v'", contact.Msg().Action())
	}
}

func (b BoltFeedRepository) saveMessage(bucket *bbolt.Bucket, msg message.Message) error {
	key := messageKey(msg.Sequence())
	value := []byte(msg.Id().String())

	if err := bucket.Put(key, value); err != nil {
		return errors.Wrap(err, "writing to the bucket failed")
	}

	if err := b.messageRepository.Put(msg); err != nil {
		return errors.Wrap(err, "message repository put failed")
	}

	if err := b.receiveLog.Put(msg); err != nil {
		return errors.Wrap(err, "receive log put failed")
	}

	return nil
}

func messageKey(seq message.Sequence) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(seq.Int()))
	return buf
}

func createFeedBucket(tx *bbolt.Tx, ref refs.Feed) (bucket *bbolt.Bucket, err error) {
	return createBucket(tx, feedBucketPath(ref))
}

func getFeedBucket(tx *bbolt.Tx, ref refs.Feed) (*bbolt.Bucket, error) {
	return getBucket(tx, feedBucketPath(ref))
}

func feedBucketPath(ref refs.Feed) []bucketName {
	return []bucketName{
		bucketName("feeds"),
		bucketName(ref.String()),
	}
}
