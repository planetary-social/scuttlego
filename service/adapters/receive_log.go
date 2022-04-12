package adapters

import (
	"encoding/binary"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"go.etcd.io/bbolt"
)

type ReceiveLogReadRepository struct {
	db         *bbolt.DB
	identifier RawMessageIdentifier
}

func NewReceiveLogReadRepository(db *bbolt.DB, identifier RawMessageIdentifier) *ReceiveLogReadRepository {
	return &ReceiveLogReadRepository{db: db, identifier: identifier}
}

func (r ReceiveLogReadRepository) Next(lastSeq uint64) ([]message.Message, error) {
	var result []message.Message

	if err := r.db.View(func(tx *bbolt.Tx) error {
		bucket, err := getReceiveLogBucket(tx)
		if err != nil {
			return errors.Wrap(err, "could not create a bucket")
		}

		if bucket == nil {
			return nil
		}

		key, value := bucket.Cursor().Seek(itob(lastSeq))
		if key == nil {
			return nil
		}

		msg, err := r.identifier.IdentifyRawMessage(message.NewRawMessage(value))
		if err != nil {
			return errors.Wrap(err, "failed to identify raw message")
		}

		result = []message.Message{msg}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}

type ReceiveLogRepository struct {
	tx         *bbolt.Tx
	identifier RawMessageIdentifier
}

func NewReceiveLogRepository(tx *bbolt.Tx, identifier RawMessageIdentifier) *ReceiveLogRepository {
	return &ReceiveLogRepository{tx: tx, identifier: identifier}
}

func (r ReceiveLogRepository) Put(msg message.Message) error {
	bucket, err := createReceiveLogBucket(r.tx)
	if err != nil {
		return errors.Wrap(err, "could not create a bucket")
	}

	return r.saveMessage(bucket, msg)
}

func (r ReceiveLogRepository) saveMessage(bucket *bbolt.Bucket, msg message.Message) error {
	seq, err := bucket.NextSequence()
	if err != nil {
		return errors.Wrap(err, "could not get the next sequence")
	}

	key := itob(seq)

	return bucket.Put(key, msg.Raw().Bytes())
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func createReceiveLogBucket(tx *bbolt.Tx) (bucket *bbolt.Bucket, err error) {
	bucketNames := receiveLogBucketPath()
	if len(bucketNames) == 0 {
		return nil, errors.New("path func returned an empty slice")
	}

	return createBucket(tx, bucketNames)
}

func getReceiveLogBucket(tx *bbolt.Tx) (*bbolt.Bucket, error) {
	bucketNames := receiveLogBucketPath()
	if len(bucketNames) == 0 {
		return nil, errors.New("path func returned an empty slice")
	}

	return getBucket(tx, bucketNames), nil
}

func receiveLogBucketPath() []bucketName {
	return []bucketName{
		bucketName("receive_log"),
	}
}
