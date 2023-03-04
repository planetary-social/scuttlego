package notx

import (
	"context"
	"time"

	"github.com/planetary-social/scuttlego/logging"
)

type NoTxFeedWantListRepository struct {
	transaction TransactionProvider
	logger      logging.Logger
}

func NewNoTxFeedWantListRepository(transaction TransactionProvider, logger logging.Logger) *NoTxFeedWantListRepository {
	return &NoTxFeedWantListRepository{
		transaction: transaction,
		logger:      logger.New("no_tx_feed_want_list_repository"),
	}
}

func (r NoTxFeedWantListRepository) CleanupLoop(ctx context.Context) error {
	for {
		if err := r.transaction.Update(func(adapters TxAdapters) error {
			return adapters.FeedWantListRepository.Cleanup()
		}); err != nil {
			r.logger.Error().WithError(err).Message("transaction failed")
		}

		select {
		case <-time.After(cleanupWantListsEvery):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
