package notx

//import (
//	"github.com/boreq/errors"
//	"github.com/planetary-social/scuttlego/service/domain/replication"
//	"go.etcd.io/bbolt"
//)
//
//type ReadWantedFeedsRepository struct {
//	db      *bbolt.DB
//	factory TxRepositoriesFactory
//}
//
//func NewReadWantedFeedsRepository(db *bbolt.DB, factory TxRepositoriesFactory) *ReadWantedFeedsRepository {
//	return &ReadWantedFeedsRepository{db: db, factory: factory}
//}
//
//func (b ReadWantedFeedsRepository) GetWantedFeeds() (replication.WantedFeeds, error) {
//	var result replication.WantedFeeds
//
//	if err := b.db.Update(func(tx *bbolt.Tx) error {
//		r, err := b.factory(tx)
//		if err != nil {
//			return errors.Wrap(err, "could not call the factory")
//		}
//
//		wantedFeeds, err := r.WantedFeeds.GetWantedFeeds()
//		if err != nil {
//			return errors.Wrap(err, "could not get wanted feeds")
//		}
//
//		result = wantedFeeds
//
//		return nil
//	}); err != nil {
//		return replication.WantedFeeds{}, errors.Wrap(err, "transaction failed")
//	}
//
//	return result, nil
//}
