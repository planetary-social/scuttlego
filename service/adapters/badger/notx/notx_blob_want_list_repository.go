package notx

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type NoTxBlobWantListRepository struct {
	transaction TransactionProvider
	logger      logging.Logger
}

func NewNoTxBlobWantListRepository(transaction TransactionProvider, logger logging.Logger) *NoTxBlobWantListRepository {
	return &NoTxBlobWantListRepository{
		transaction: transaction,
		logger:      logger.New("no_tx_blob_want_list_repository"),
	}
}

func (b NoTxBlobWantListRepository) GetWantedBlobs() ([]refs.Blob, error) {
	var result []refs.Blob

	if err := b.transaction.View(func(adapters TxAdapters) error {
		tmp, err := adapters.BlobWantListRepository.List()
		if err != nil {
			return errors.Wrap(err, "could not get blobs from tx repo")
		}
		result = tmp
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}

func (b NoTxBlobWantListRepository) Contains(id refs.Blob) (bool, error) {
	var result bool

	if err := b.transaction.View(func(adapters TxAdapters) error {
		contains, err := adapters.BlobWantListRepository.Contains(id)
		if err != nil {
			return errors.Wrap(err, "could not call contains on tx repo")
		}

		result = contains
		return nil
	}); err != nil {
		return false, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}

func (b NoTxBlobWantListRepository) Delete(id refs.Blob) error {
	if err := b.transaction.Update(func(adapters TxAdapters) error {
		return adapters.BlobWantListRepository.Delete(id)
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}

const cleanupWantListsEvery = 1 * time.Minute

func (r NoTxBlobWantListRepository) CleanupLoop(ctx context.Context) error {
	for {
		if err := r.transaction.Update(func(adapters TxAdapters) error {
			return adapters.BlobWantListRepository.Cleanup()
		}); err != nil {
			r.logger.WithError(err).Error("transaction failed")
		}

		select {
		case <-time.After(cleanupWantListsEvery):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
