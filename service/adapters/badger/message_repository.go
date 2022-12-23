package badger

import (
	"encoding/binary"

	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/planetary-social/scuttlego/service/adapters/badger/utils"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

const (
	messageRepositoryMetaBucketCountKey = "message_count"
)

type RawMessageIdentifier interface {
	IdentifyRawMessage(raw message.RawMessage) (message.Message, error)
}

type MessageRepository struct {
	tx         *badger.Txn
	identifier RawMessageIdentifier
}

func NewMessageRepository(
	tx *badger.Txn,
	identifier RawMessageIdentifier,
) *MessageRepository {
	return &MessageRepository{
		tx:         tx,
		identifier: identifier,
	}
}

func (r MessageRepository) Put(msg message.Message) error {
	bucket, err := r.createMessageBucket()
	if err != nil {
		return errors.Wrap(err, "could not create the bucket")
	}

	key := r.messageKey(msg.Id())

	if _, err := bucket.Get(key); err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			// todo test
			counter, err := r.getMessageCounter()
			if err != nil {
				return errors.Wrap(err, "could not get the counter")
			}

			if err := counter.Increment(); err != nil {
				return errors.Wrap(err, "error incrementing count")
			}
		} else {
			return errors.Wrap(err, "error checking if the message exists")
		}
	}

	if err := bucket.Set(key, msg.Raw().Bytes()); err != nil {
		return errors.Wrap(err, "bucket put failed")
	}

	return nil
}

func (r MessageRepository) Get(id refs.Message) (message.Message, error) {
	bucket, err := r.createMessageBucket()
	if err != nil {
		return message.Message{}, errors.Wrap(err, "could not get the bucket")
	}

	item, err := bucket.Get(r.messageKey(id))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return message.Message{}, errors.Wrap(err, "message not found")
		}

		return message.Message{}, errors.Wrap(err, "error getting message")
	}

	var result message.Message

	if err := item.Value(func(value []byte) error {
		rawMsg, err := message.NewRawMessage(value) // todo explicit copy? it is copied in constructor
		if err != nil {
			return errors.Wrap(err, "could not create a raw message")
		}

		msg, err := r.identifier.IdentifyRawMessage(rawMsg)
		if err != nil {
			return errors.Wrap(err, "could not identify the raw message")
		}

		result = msg
		return nil
	}); err != nil {
		return message.Message{}, errors.Wrap(err, "error getting the message value")
	}

	return result, nil
}

func (r MessageRepository) Delete(id refs.Message) error {
	bucket, err := r.createMessageBucket()
	if err != nil {
		return errors.Wrap(err, "could not get the bucket")
	}

	key := r.messageKey(id)

	if _, err := bucket.Get(key); err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		return errors.Wrap(err, "error checking if the message exists")
	}

	// todo test
	counter, err := r.getMessageCounter()
	if err != nil {
		return errors.Wrap(err, "could not get the counter")
	}

	if err := counter.Decrement(); err != nil {
		return errors.Wrap(err, "error decrementing counter")
	}

	if err := bucket.Delete(key); err != nil {
		return errors.Wrap(err, "bucket delete failed")
	}

	return nil
}

func (r MessageRepository) Count() (int, error) {
	counter, err := r.getMessageCounter()
	if err != nil {
		return 0, errors.Wrap(err, "could not get the counter")
	}

	count, err := counter.Get()
	if err != nil {
		return 0, errors.Wrap(err, "error getting count")
	}

	return int(count), nil
}

func (r MessageRepository) marshalCount(count uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, count)
	return b
}

func (r MessageRepository) unmarshalCount(b []byte) (uint64, error) {
	if len(b) != 8 {
		return 0, errors.New("invalid length")
	}
	return binary.LittleEndian.Uint64(b), nil
}

func (r MessageRepository) messageKey(id refs.Message) []byte {
	return []byte(id.String())
}

func (r MessageRepository) getMessageCounter() (utils.Counter, error) {
	b, err := r.createMetaBucket()
	if err != nil {
		return utils.Counter{}, errors.Wrap(err, "error creating meta bucket")
	}

	return utils.NewCounter(b, utils.MustNewKeyComponent([]byte("message_count")))
}

func (r MessageRepository) createMessageBucket() (utils.Bucket, error) {
	return utils.NewBucket(
		r.tx,
		utils.MustNewKey(
			utils.MustNewKeyComponent([]byte("messages")),
			utils.MustNewKeyComponent([]byte("entries")),
		),
	)
}

func (r MessageRepository) createMetaBucket() (utils.Bucket, error) {
	return utils.NewBucket(
		r.tx,
		utils.MustNewKey(
			utils.MustNewKeyComponent([]byte("messages")),
			utils.MustNewKeyComponent([]byte("meta")),
		),
	)
}
