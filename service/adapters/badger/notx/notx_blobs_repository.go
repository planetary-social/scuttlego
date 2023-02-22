package notx

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type NoTxBlobsRepository struct {
	transaction TransactionProvider
}

func NewNoTxBlobsRepository(transaction TransactionProvider) *NoTxBlobsRepository {
	return &NoTxBlobsRepository{transaction: transaction}
}

func (b NoTxBlobsRepository) GetFeedBlobs(id refs.Feed) ([]replication.MessageBlobs, error) {
	var result []replication.MessageBlobs

	if err := b.transaction.View(func(adapters TxAdapters) error {
		messages, err := adapters.FeedRepository.GetMessages(id, nil, nil)
		if err != nil {
			return errors.Wrap(err, "error getting messages")
		}

		for _, msg := range messages {
			blobs, err := adapters.BlobRepository.ListBlobs(msg.Id())
			if err != nil {
				return errors.Wrapf(err, "error getting blobs for message '%s'", msg.Id())
			}

			result = append(result, replication.MessageBlobs{
				Message: msg,
				Blobs:   blobs,
			})
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}
