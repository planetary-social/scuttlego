package notx

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type NoTxBlobWantListRepository struct {
	transaction *TransactionProvider
}

func NewNoTxBlobWantListRepository(transaction *TransactionProvider) *NoTxBlobWantListRepository {
	return &NoTxBlobWantListRepository{transaction: transaction}
}

func (b NoTxBlobWantListRepository) List() (blobs.WantList, error) {
	var result []blobs.WantedBlob

	if err := b.transaction.View(func(adapters TxAdapters) error {
		list, err := adapters.BlobWantListRepository.List()
		if err != nil {
			return errors.Wrap(err, "could not get blobs")
		}

		for _, v := range list {
			result = append(result, blobs.WantedBlob{
				Id:       v,
				Distance: blobs.NewWantDistanceLocal(),
			})
		}

		return nil
	}); err != nil {
		return blobs.WantList{}, errors.Wrap(err, "transaction failed")
	}

	return blobs.NewWantList(result)
}

func (b NoTxBlobWantListRepository) Contains(id refs.Blob) (bool, error) {
	var result bool

	if err := b.transaction.View(func(adapters TxAdapters) error {
		contains, err := adapters.BlobWantListRepository.Contains(id)
		if err != nil {
			return errors.Wrap(err, "could not get blobs")
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
