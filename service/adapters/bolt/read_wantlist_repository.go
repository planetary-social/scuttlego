package bolt

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/blobs"
	"go.etcd.io/bbolt"
)

type ReadWantListRepository struct {
	db      *bbolt.DB
	factory TxRepositoriesFactory
}

func NewReadWantListRepository(db *bbolt.DB, factory TxRepositoriesFactory) *ReadWantListRepository {
	return &ReadWantListRepository{db: db, factory: factory}
}

func (b ReadWantListRepository) GetWantList() (blobs.WantList, error) {
	var result []blobs.WantedBlob

	if err := b.db.View(func(tx *bbolt.Tx) error {
		r, err := b.factory(tx)
		if err != nil {
			return errors.Wrap(err, "could not call the factory")
		}

		list, err := r.Blob.List()
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
