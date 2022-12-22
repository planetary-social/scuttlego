package notx

//type ReadMessageRepository struct {
//	db      *bbolt.DB
//	factory TxRepositoriesFactory
//}
//
//func NewReadMessageRepository(db *bbolt.DB, factory TxRepositoriesFactory) *ReadMessageRepository {
//	return &ReadMessageRepository{
//		db:      db,
//		factory: factory,
//	}
//}
//
//func (r ReadMessageRepository) Count() (int, error) {
//	var result int
//
//	if err := r.db.View(func(tx *bbolt.Tx) error {
//		r, err := r.factory(tx)
//		if err != nil {
//			return errors.Wrap(err, "could not call the factory")
//		}
//
//		n, err := r.Message.Count()
//		if err != nil {
//			return errors.Wrap(err, "failed calling the repo")
//		}
//
//		result = n
//
//		return nil
//	}); err != nil {
//		return 0, errors.Wrap(err, "transaction failed")
//	}
//
//	return result, nil
//}
