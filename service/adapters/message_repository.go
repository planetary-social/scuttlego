package adapters

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"go.etcd.io/bbolt"
)

type RawMessageIdentifier interface {
	IdentifyRawMessage(raw message.RawMessage) (message.Message, error)
}

type BoltMessageRepository struct {
	tx         *bbolt.Tx
	identifier RawMessageIdentifier
}

func NewBoltMessageRepository(
	tx *bbolt.Tx,
	identifier RawMessageIdentifier,
) *BoltMessageRepository {
	return &BoltMessageRepository{
		tx:         tx,
		identifier: identifier,
	}
}

func (r BoltMessageRepository) Put(msg message.Message) error {
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

func (r BoltMessageRepository) Get(id refs.Message) (message.Message, error) {
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

func (r BoltMessageRepository) messageKey(id refs.Message) []byte {
	return []byte(id.String())
}

func (r BoltMessageRepository) createBucket() (*bbolt.Bucket, error) {
	return createBucket(r.tx, r.bucketPath())
}

func (r BoltMessageRepository) getBucket() (*bbolt.Bucket, error) {
	return getBucket(r.tx, r.bucketPath())
}

func (r BoltMessageRepository) bucketPath() []bucketName {
	return []bucketName{
		bucketName("messages"),
	}
}
