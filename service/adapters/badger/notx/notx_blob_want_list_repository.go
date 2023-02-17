package notx

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type NoTxBlobWantListRepository struct {
	transaction TransactionProvider
}

func NewNoTxBlobWantListRepository(transaction TransactionProvider) *NoTxBlobWantListRepository {
	return &NoTxBlobWantListRepository{transaction: transaction}
}

func (b NoTxBlobWantListRepository) GetWantedBlobs() ([]refs.Blob, error) {
	var result []refs.Blob

	if err := b.transaction.Update(func(adapters TxAdapters) error {
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
