package notx

import (
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"go.etcd.io/bbolt"
)

type ReadFeedRepository struct {
	transaction *TransactionProvider
}

func NewReadFeedRepository(
	transaction *TransactionProvider,
) *ReadFeedRepository {
	return &ReadFeedRepository{
		transaction: transaction,,
	}
}

func (b ReadFeedRepository) GetMessages(id refs.Feed, seq *message.Sequence, limit *int) ([]message.Message, error) {
	var result []message.Message

	if err := b.transaction.View(func(adapters TxAdapters) error {
		messages, err := adapters.Feed.GetMessages(id, seq, limit)
		if err != nil {
			return errors.Wrap(err, "failed to call the feed repository")
		}

		result = messages

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}

func (b ReadFeedRepository) Count() (int, error) {
	var result int

	if err := b.db.View(func(tx *bbolt.Tx) error {
		r, err := b.factory(tx)
		if err != nil {
			return errors.Wrap(err, "factory returned an error")
		}

		n, err := r.Feed.Count()
		if err != nil {
			return errors.Wrap(err, "failed to call the feed repository")
		}

		result = n

		return nil
	}); err != nil {
		return 0, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}

func (b ReadFeedRepository) GetMessage(feed refs.Feed, sequence message.Sequence) (message.Message, error) {
	var result message.Message

	if err := b.db.View(func(tx *bbolt.Tx) error {
		r, err := b.factory(tx)
		if err != nil {
			return errors.Wrap(err, "factory returned an error")
		}

		msg, err := r.Feed.GetMessage(feed, sequence)
		if err != nil {
			return errors.Wrap(err, "failed to call the feed repository")
		}

		result = msg

		return nil
	}); err != nil {
		return message.Message{}, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}
