package bolt

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"go.etcd.io/bbolt"
)

type BoltFeedMessagesRepository struct {
	db      *bbolt.DB
	factory TxRepositoriesFactory
}

func NewBoltFeedMessagesRepository(
	db *bbolt.DB,
	factory TxRepositoriesFactory,
) *BoltFeedMessagesRepository {
	return &BoltFeedMessagesRepository{
		db:      db,
		factory: factory,
	}
}

func (b BoltFeedMessagesRepository) GetMessages(id refs.Feed, seq *message.Sequence, limit *int) ([]message.Message, error) {
	var messages []message.Message

	if err := b.db.View(func(tx *bbolt.Tx) error {
		r, err := b.factory(tx)
		if err != nil {
			return errors.Wrap(err, "factory returned an error")
		}

		bucket, err := getFeedBucket(tx, id)
		if err != nil {
			return errors.Wrap(err, "could not get the bucket")
		}

		if bucket == nil {
			return nil
		}

		// todo not stupid implementation (with a cursor)

		if err := bucket.ForEach(func(key, value []byte) error {
			msgId, err := refs.NewMessage(string(value))
			if err != nil {
				return errors.Wrap(err, "failed to create a message ref")
			}

			msg, err := r.Message.Get(msgId)
			if err != nil {
				return errors.Wrap(err, "failed to get the message")
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
