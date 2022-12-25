package notx

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type NoTxFeedRepository struct {
	transaction *TransactionProvider
}

func NewNoTxFeedRepository(
	transaction *TransactionProvider,
) *NoTxFeedRepository {
	return &NoTxFeedRepository{
		transaction: transaction,
	}
}

func (b NoTxFeedRepository) GetMessages(id refs.Feed, seq *message.Sequence, limit *int) ([]message.Message, error) {
	var result []message.Message

	if err := b.transaction.View(func(adapters TxAdapters) error {
		messages, err := adapters.FeedRepository.GetMessages(id, seq, limit)
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

func (b NoTxFeedRepository) Count() (int, error) {
	var result int

	if err := b.transaction.View(func(adapters TxAdapters) error {
		tmp, err := adapters.FeedRepository.Count()
		if err != nil {
			return errors.Wrap(err, "failed to call the feed repository")
		}

		result = tmp
		return nil
	}); err != nil {
		return 0, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}

func (b NoTxFeedRepository) GetMessage(feed refs.Feed, sequence message.Sequence) (message.Message, error) {
	var result message.Message

	if err := b.transaction.View(func(adapters TxAdapters) error {
		tmp, err := adapters.FeedRepository.GetMessage(feed, sequence)
		if err != nil {
			return errors.Wrap(err, "failed to call the feed repository")
		}

		result = tmp
		return nil
	}); err != nil {
		return message.Message{}, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}
