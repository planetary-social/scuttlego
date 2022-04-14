package bolt

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"go.etcd.io/bbolt"
)

type RawMessageIdentifier interface {
	IdentifyRawMessage(raw message.RawMessage) (message.Message, error)
}

type MessageRepository struct {
	tx         *bbolt.Tx
	identifier RawMessageIdentifier
}

func NewMessageRepository(
	tx *bbolt.Tx,
	identifier RawMessageIdentifier,
) *MessageRepository {
	return &MessageRepository{
		tx:         tx,
		identifier: identifier,
	}
}

func (r MessageRepository) Put(msg message.Message) error {
	bucket, err := r.createBucket()
	if err != nil {
		return errors.Wrap(err, "could not create the bucket")
	}

	key := r.messageKey(msg.Id())

	if err := bucket.Put(key, msg.Raw().Bytes()); err != nil {
		return errors.Wrap(err, "bucket put failed")
	}

	return nil
}

func (r MessageRepository) Get(id refs.Message) (message.Message, error) {
	bucket, err := r.getBucket()
	if err != nil {
		return message.Message{}, errors.Wrap(err, "could not get the bucket")
	}

	if bucket == nil {
		return message.Message{}, errors.New("message not found")
	}

	value := bucket.Get(r.messageKey(id))

	if value == nil {
		return message.Message{}, errors.New("message not found")
	}

	rawMsg := message.NewRawMessage(value)

	msg, err := r.identifier.IdentifyRawMessage(rawMsg)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "could not identify the raw message")
	}

	return msg, nil
}

func (r MessageRepository) Count() (int, error) {
	bucket, err := r.getBucket()
	if err != nil {
		return 0, errors.Wrap(err, "could not get the bucket")
	}

	if bucket == nil {
		return 0, nil
	}

	// this works only because there are no nested buckets inside of the message bucket
	return bucket.Stats().KeyN, nil
}

func (r MessageRepository) messageKey(id refs.Message) []byte {
	return []byte(id.String())
}

func (r MessageRepository) createBucket() (*bbolt.Bucket, error) {
	return createBucket(r.tx, r.bucketPath())
}

func (r MessageRepository) getBucket() (*bbolt.Bucket, error) {
	return getBucket(r.tx, r.bucketPath())
}

func (r MessageRepository) bucketPath() []bucketName {
	return []bucketName{
		bucketName("messages"),
	}
}

type ReadMessageRepository struct {
	db      *bbolt.DB
	factory TxRepositoriesFactory
}

func NewReadMessageRepository(db *bbolt.DB, factory TxRepositoriesFactory) *ReadMessageRepository {
	return &ReadMessageRepository{
		db:      db,
		factory: factory,
	}
}

func (r ReadMessageRepository) Count() (int, error) {
	var result int

	if err := r.db.View(func(tx *bbolt.Tx) error {
		r, err := r.factory(tx)
		if err != nil {
			return errors.Wrap(err, "could not call the factory")
		}

		n, err := r.Message.Count()
		if err != nil {
			return errors.Wrap(err, "failed calling the repo")
		}

		result = n

		return nil
	}); err != nil {
		return 0, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}
