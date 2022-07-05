package bolt

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"go.etcd.io/bbolt"
)

type NoTxWantListRepository struct {
	db      *bbolt.DB
	factory TxRepositoriesFactory
}

func NewNoTxWantListRepository(db *bbolt.DB, factory TxRepositoriesFactory) *NoTxWantListRepository {
	return &NoTxWantListRepository{db: db, factory: factory}
}

func (b NoTxWantListRepository) GetWantList() (blobs.WantList, error) {
	var result []blobs.WantedBlob

	if err := b.db.Batch(func(tx *bbolt.Tx) error {
		r, err := b.factory(tx)
		if err != nil {
			return errors.Wrap(err, "could not call the factory")
		}

		list, err := r.WantList.List()
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

func (b NoTxWantListRepository) WantListContains(id refs.Blob) (bool, error) {
	var result bool

	if err := b.db.View(func(tx *bbolt.Tx) error {
		r, err := b.factory(tx)
		if err != nil {
			return errors.Wrap(err, "could not call the factory")
		}

		contains, err := r.WantList.WantListContains(id)
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

func (b NoTxWantListRepository) DeleteFromWantList(id refs.Blob) error {
	if err := b.db.Batch(func(tx *bbolt.Tx) error {
		r, err := b.factory(tx)
		if err != nil {
			return errors.Wrap(err, "could not call the factory")
		}

		return r.WantList.DeleteFromWantList(id)
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
