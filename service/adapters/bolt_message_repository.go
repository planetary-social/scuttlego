package adapters

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"go.etcd.io/bbolt"
)

type BoltMessageRepository struct {
	db         *bbolt.DB
	identifier RawMessageIdentifier
}

func NewBoltMessageRepository(
	db *bbolt.DB,
	identifier RawMessageIdentifier,
) *BoltMessageRepository {
	return &BoltMessageRepository{
		db:         db,
		identifier: identifier,
	}
}

func (b BoltMessageRepository) GetMessages(id refs.Feed, seq *message.Sequence, limit *int) ([]message.Message, error) {
	var messages []message.Message

	if err := b.db.View(func(tx *bbolt.Tx) error {
		bucket, err := getFeedBucket(tx, id)
		if err != nil {
			return errors.Wrap(err, "could not get the bucket")
		}

		if bucket == nil {
			return nil
		}

		// todo not stupid implementation (with a cursor)

		if err := bucket.ForEach(func(key, value []byte) error {
			rawMsg := message.NewRawMessage(value)

			msg, err := b.identifier.IdentifyRawMessage(rawMsg)
			if err != nil {
				return errors.Wrap(err, "could not identify the raw message")
			}

			if (limit == nil || len(messages) < *limit) && (seq == nil || !seq.ComesAfter(msg.Sequence())) {
				messages = append(messages, msg)
			}

			return nil
		}); err != nil {
			return errors.Wrap(err, "failed to iterate")
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return messages, nil
}
