package bolt

import (
	"encoding/binary"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"go.etcd.io/bbolt"
)

type ReceiveLogRepository struct {
	tx                *bbolt.Tx
	messageRepository *MessageRepository
}

func NewReceiveLogRepository(tx *bbolt.Tx, messageRepository *MessageRepository) *ReceiveLogRepository {
	return &ReceiveLogRepository{
		tx:                tx,
		messageRepository: messageRepository,
	}
}

func (r ReceiveLogRepository) Put(id refs.Message) error {
	bucket, err := r.createReceiveLogBucket(r.tx)
	if err != nil {
		return errors.Wrap(err, "could not create a bucket")
	}

	return r.saveMessage(bucket, id)
}

func (r ReceiveLogRepository) Next(lastSeq uint64) ([]message.Message, error) {
	bucket, err := r.getReceiveLogBucket(r.tx)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a bucket")
	}

	if bucket == nil {
		return nil, nil
	}

	key, value := bucket.Cursor().Seek(itob(lastSeq))
	if key == nil {
		return nil, nil
	}

	id, err := refs.NewMessage(string(value))
	if err != nil {
		return nil, errors.New("could not create a message ref")
	}

	msg, err := r.messageRepository.Get(id)
	if err != nil {
		return nil, errors.New("could not get the message")
	}

	return []message.Message{msg}, nil
}

func (r ReceiveLogRepository) saveMessage(bucket *bbolt.Bucket, id refs.Message) error {
	seq, err := bucket.NextSequence()
	if err != nil {
		return errors.Wrap(err, "could not get the next sequence")
	}

	key := itob(seq)

	return bucket.Put(key, []byte(id.String()))
}

func (r ReceiveLogRepository) createReceiveLogBucket(tx *bbolt.Tx) (*bbolt.Bucket, error) {
	return createBucket(tx, r.receiveLogBucketPath())
}

func (r ReceiveLogRepository) getReceiveLogBucket(tx *bbolt.Tx) (*bbolt.Bucket, error) {
	return getBucket(tx, r.receiveLogBucketPath())
}

func (r ReceiveLogRepository) receiveLogBucketPath() []bucketName {
	return []bucketName{
		bucketName("receive_log"),
	}
}

type ReceiveLogReadRepository struct {
	db         *bbolt.DB
	identifier RawMessageIdentifier
}

func NewReceiveLogReadRepository(db *bbolt.DB, identifier RawMessageIdentifier) *ReceiveLogReadRepository {
	return &ReceiveLogReadRepository{
		db:         db,
		identifier: identifier,
	}
}

func (r ReceiveLogReadRepository) Next(lastSeq uint64) ([]message.Message, error) {
	var result []message.Message

	if err := r.db.View(func(tx *bbolt.Tx) error {
		messageRepository := NewMessageRepository(tx, r.identifier)

		repository := NewReceiveLogRepository(tx, messageRepository)
		msgs, err := repository.Next(lastSeq)
		if err != nil {
			return errors.Wrap(err, "failed to call the repository")
		}

		result = msgs
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
