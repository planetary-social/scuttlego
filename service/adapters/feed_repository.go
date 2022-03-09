package adapters

import (
	"encoding/binary"
	"fmt"

	feeds2 "github.com/planetary-social/go-ssb/service/domain/feeds"
	"github.com/planetary-social/go-ssb/service/domain/feeds/content"
	"github.com/planetary-social/go-ssb/service/domain/feeds/formats"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/replication"

	"github.com/boreq/errors"
	"go.etcd.io/bbolt"
)

type RawMessageIdentifier interface {
	IdentifyRawMessage(raw message.RawMessage) (message.Message, error)
}

type BoltFeedRepository struct {
	tx                *bbolt.Tx
	identifier        RawMessageIdentifier
	graph             *SocialGraphRepository
	formatScuttlebutt *formats.Scuttlebutt
}

func NewBoltFeedRepository(
	tx *bbolt.Tx,
	identifier RawMessageIdentifier,
	graph *SocialGraphRepository,
	formatScuttlebutt *formats.Scuttlebutt,
) *BoltFeedRepository {
	return &BoltFeedRepository{
		tx:                tx,
		identifier:        identifier,
		graph:             graph,
		formatScuttlebutt: formatScuttlebutt,
	}
}

func (b BoltFeedRepository) UpdateFeed(ref refs.Feed, f func(feed *feeds2.Feed) (*feeds2.Feed, error)) error {
	feed, err := b.loadFeed(ref)
	if err != nil {
		return errors.Wrap(err, "could not load a feed")
	}

	if feed == nil {
		feed = feeds2.NewFeed(b.formatScuttlebutt)
	}

	feed, err = f(feed)
	if err != nil {
		return errors.Wrap(err, "provided function returned an error")
	}

	return b.saveFeed(ref, feed)
}
func (b BoltFeedRepository) GetFeed(ref refs.Feed) (*feeds2.Feed, error) {
	f, err := b.loadFeed(ref)
	if err != nil {
		return nil, errors.Wrap(err, "loading failed")
	}

	if f == nil {
		return nil, replication.ErrFeedNotFound
	}

	return f, nil
}

func (b BoltFeedRepository) loadFeed(ref refs.Feed) (*feeds2.Feed, error) {
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

	rawMsg := message.NewRawMessage(value)

	msg, err := b.identifier.IdentifyRawMessage(rawMsg)
	if err != nil {
		return nil, errors.Wrap(err, "could not identify the raw message")
	}

	feed, err := feeds2.NewFeedFromHistory(msg, b.formatScuttlebutt)
	if err != nil {
		return nil, errors.Wrap(err, "could not recreate a feed from history")
	}

	return feed, nil
}

func (b BoltFeedRepository) saveFeed(ref refs.Feed, feed *feeds2.Feed) error {
	msgs, contacts := feed.PopForPersisting()

	if len(msgs) != 0 {
		bucket, err := b.createFeedBucket(ref)
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

func (b BoltFeedRepository) saveContact(contact feeds2.ContactToSave) error {
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
	key := b.messageKey(msg)
	return bucket.Put(key, msg.Raw().Bytes())
}

func (b BoltFeedRepository) messageKey(msg message.Message) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(msg.Sequence().Int()))
	return buf
}

func (b *BoltFeedRepository) createFeedBucket(ref refs.Feed) (bucket *bbolt.Bucket, err error) {
	bucketNames := b.pathFunc(ref)
	if len(bucketNames) == 0 {
		return nil, errors.New("path func returned an empty slice")
	}

	return createBucket(b.tx, bucketNames)
}

func (b *BoltFeedRepository) getFeedBucket(ref refs.Feed) (*bbolt.Bucket, error) {
	bucketNames := b.pathFunc(ref)
	if len(bucketNames) == 0 {
		return nil, errors.New("path func returned an empty slice")
	}

	return getBucket(b.tx, bucketNames), nil
}

func (b *BoltFeedRepository) pathFunc(ref refs.Feed) []BucketName {
	return []BucketName{
		BucketName("feeds"),
		BucketName(ref.String()),
	}
}
