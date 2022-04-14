package bolt

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"go.etcd.io/bbolt"
)

type ReadFeedRepository struct {
	db      *bbolt.DB
	factory TxRepositoriesFactory
}

func NewReadFeedRepository(
	db *bbolt.DB,
	factory TxRepositoriesFactory,
) *ReadFeedRepository {
	return &ReadFeedRepository{
		db:      db,
		factory: factory,
	}
}

func (b ReadFeedRepository) GetMessages(id refs.Feed, seq *message.Sequence, limit *int) ([]message.Message, error) {
	var result []message.Message

	if err := b.db.View(func(tx *bbolt.Tx) error {
		r, err := b.factory(tx)
		if err != nil {
			return errors.Wrap(err, "factory returned an error")
		}

		messages, err := r.Feed.GetMessages(id, seq, limit)
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
