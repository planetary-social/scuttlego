package adapters

import (
	"encoding/binary"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/refs"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
	"github.com/planetary-social/go-ssb/scuttlebutt/replication"
	bolt "go.etcd.io/bbolt"
)

type RawMessageIdentifier interface {
	IdentifyRawMessage(raw message.RawMessage) (message.Message, error)
}

type BucketName []byte

type BoltFeedStorage struct {
	db         *bolt.DB
	identifier RawMessageIdentifier
}

func NewBoltFeedStorage(db *bolt.DB, identifier RawMessageIdentifier) *BoltFeedStorage {
	return &BoltFeedStorage{
		db:         db,
		identifier: identifier,
	}
}

func (b BoltFeedStorage) UpdateFeed(ref refs.Feed, f func(feed *feeds.Feed) (*feeds.Feed, error)) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		feed, err := b.loadFeed(tx, ref)
		if err != nil {
			return errors.Wrap(err, "could not load a feed")
		}

		feed, err = f(feed)
		if err != nil {
			return errors.Wrap(err, "provided function returned an error")
		}

		return b.saveFeed(tx, feed)
	})
}
func (b BoltFeedStorage) GetFeed(ref refs.Feed) (*feeds.Feed, error) {
	var f *feeds.Feed

	if err := b.db.View(func(tx *bolt.Tx) error {
		var err error
		f, err = b.loadFeed(tx, ref)
		if err != nil {
			return errors.Wrap(err, "loading failed")
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	if f == nil {
		return nil, replication.ErrFeedNotFound
	}

	return f, nil
}

func (b BoltFeedStorage) loadFeed(tx *bolt.Tx, ref refs.Feed) (*feeds.Feed, error) {
	bucket, err := b.getFeedBucket(tx, ref)
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

	feed, err := feeds.NewFeedFromHistory(msg)
	if err != nil {
		return nil, errors.Wrap(err, "could not recreate a feed from history")
	}

	return feed, nil
}

func (b BoltFeedStorage) saveFeed(tx *bolt.Tx, feed *feeds.Feed) error {
	msgs, _ := feed.PopForPersisting()
	// todo save contacts

	if len(msgs) == 0 {
		return nil
	}

	bucket, err := b.createFeedBucket(tx, feed.Ref())
	if err != nil {
		return errors.Wrap(err, "could not create the bucket")
	}

	for _, msg := range msgs {
		if err := b.saveMessage(bucket, msg); err != nil {
			return errors.Wrap(err, "could not save a message")
		}
	}

	return nil
}

func (b BoltFeedStorage) saveMessage(bucket *bolt.Bucket, msg message.Message) error {
	key := b.messageKey(msg)
	return bucket.Put(key, msg.Raw().Bytes())
}

func (b BoltFeedStorage) messageKey(msg message.Message) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(msg.Sequence().Int()))
	return buf
}

func (b *BoltFeedStorage) createFeedBucket(tx *bolt.Tx, ref refs.Feed) (bucket *bolt.Bucket, err error) {
	bucketNames := b.pathFunc(ref)
	if len(bucketNames) == 0 {
		return nil, errors.New("path func returned an empty slice")
	}

	return createBucket(tx, bucketNames)
}

func (b *BoltFeedStorage) getFeedBucket(tx *bolt.Tx, ref refs.Feed) (*bolt.Bucket, error) {
	bucketNames := b.pathFunc(ref)
	if len(bucketNames) == 0 {
		return nil, errors.New("path func returned an empty slice")
	}

	return getBucket(tx, bucketNames), nil
}

func (b *BoltFeedStorage) pathFunc(ref refs.Feed) []BucketName {
	return []BucketName{
		BucketName("feeds"),
		BucketName(ref.String()),
	}
}

func getBucket(tx *bolt.Tx, bucketNames []BucketName) *bolt.Bucket {
	bucket := tx.Bucket(bucketNames[0])

	if bucket == nil {
		return nil
	}

	for i := 1; i < len(bucketNames); i++ {
		bucket = bucket.Bucket(bucketNames[i])
		if bucket == nil {
			return nil
		}
	}

	return bucket
}

func createBucket(tx *bolt.Tx, bucketNames []BucketName) (bucket *bolt.Bucket, err error) {
	bucket, err = tx.CreateBucketIfNotExists(bucketNames[0])
	if err != nil {
		return nil, errors.Wrap(err, "could not create a bucket")
	}

	for i := 1; i < len(bucketNames); i++ {
		bucket, err = bucket.CreateBucketIfNotExists(bucketNames[i])
		if err != nil {
			return nil, errors.Wrap(err, "could not create a bucket")
		}
	}

	return bucket, nil
}
