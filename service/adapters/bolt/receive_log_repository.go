package bolt

import (
	"encoding/binary"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
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

func (r ReceiveLogRepository) Get(startSeq int, limit int) ([]message.Message, error) {
	if startSeq < 0 {
		return nil, errors.New("start seq can't be negative")
	}

	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}

	bucket, err := r.getReceiveLogBucket(r.tx)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a bucket")
	}

	if bucket == nil {
		return nil, nil
	}

	var result []message.Message

	c := bucket.Cursor()
	for key, value := c.Seek(itob(uint64(startSeq))); key != nil; key, value = c.Next() {
		msg, err := r.loadMessage(value)
		if err != nil {
			return nil, errors.New("could not load a message")
		}

		result = append(result, msg)

		if len(result) >= limit {
			break
		}
	}

	return result, nil
}

func (r ReceiveLogRepository) saveMessage(bucket *bbolt.Bucket, id refs.Message) error {
	seq, err := bucket.NextSequence()
	if err != nil {
		return errors.Wrap(err, "could not get the next sequence")
	}

	key := itob(seq - 1) // NextSequence starts with 1 while our log is 0 indexed

	return bucket.Put(key, []byte(id.String()))
}

func (r ReceiveLogRepository) loadMessage(value []byte) (message.Message, error) {
	id, err := refs.NewMessage(string(value))
	if err != nil {
		return message.Message{}, errors.New("could not create a message ref")
	}

	msg, err := r.messageRepository.Get(id)
	if err != nil {
		return message.Message{}, errors.New("could not get the message")
	}

	return msg, nil
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

type ReadReceiveLogRepository struct {
	db      *bbolt.DB
	factory TxRepositoriesFactory
}

func NewReadReceiveLogRepository(db *bbolt.DB, factory TxRepositoriesFactory) *ReadReceiveLogRepository {
	return &ReadReceiveLogRepository{
		db:      db,
		factory: factory,
	}
}

func (r ReadReceiveLogRepository) Get(startSeq int, limit int) ([]message.Message, error) {
	var result []message.Message

	if err := r.db.View(func(tx *bbolt.Tx) error {
		r, err := r.factory(tx)
		if err != nil {
			return errors.Wrap(err, "could not call the factory")
		}

		msgs, err := r.ReceiveLog.Get(startSeq, limit)
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
