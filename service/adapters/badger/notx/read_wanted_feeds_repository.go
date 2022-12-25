package notx

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/replication"
)

type NoTxWantedFeedsRepository struct {
	transaction TransactionProvider
}

func NewNoTxWantedFeedsRepository(transaction TransactionProvider) *NoTxWantedFeedsRepository {
	return &NoTxWantedFeedsRepository{transaction: transaction}
}

func (b NoTxWantedFeedsRepository) GetWantedFeeds() (replication.WantedFeeds, error) {
	var result replication.WantedFeeds

	if err := b.transaction.View(func(adapters TxAdapters) error {
		tmp, err := adapters.WantedFeedsRepository.GetWantedFeeds()
		if err != nil {
			return errors.Wrap(err, "could not get wanted feeds")
		}

		result = tmp
		return nil
	}); err != nil {
		return replication.WantedFeeds{}, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}
