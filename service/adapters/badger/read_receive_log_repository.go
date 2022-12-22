package badger

//type ReadReceiveLogRepository struct {
//	db      *bbolt.DB
//	factory TxRepositoriesFactory
//}
//
//func NewReadReceiveLogRepository(db *bbolt.DB, factory TxRepositoriesFactory) *ReadReceiveLogRepository {
//	return &ReadReceiveLogRepository{
//		db:      db,
//		factory: factory,
//	}
//}
//
//func (r ReadReceiveLogRepository) List(startSeq common.ReceiveLogSequence, limit int) ([]queries.LogMessage, error) {
//	var result []queries.LogMessage
//
//	if err := r.db.View(func(tx *bbolt.Tx) error {
//		r, err := r.factory(tx)
//		if err != nil {
//			return errors.Wrap(err, "could not call the factory")
//		}
//
//		msgs, err := r.ReceiveLog.List(startSeq, limit)
//		if err != nil {
//			return errors.Wrap(err, "failed to call the repository")
//		}
//
//		result = msgs
//		return nil
//	}); err != nil {
//		return nil, errors.Wrap(err, "transaction failed")
//	}
//
//	return result, nil
//}
//
//func (r ReadReceiveLogRepository) GetMessage(seq common.ReceiveLogSequence) (message.Message, error) {
//	var result message.Message
//
//	if err := r.db.View(func(tx *bbolt.Tx) error {
//		r, err := r.factory(tx)
//		if err != nil {
//			return errors.Wrap(err, "could not call the factory")
//		}
//
//		msg, err := r.ReceiveLog.GetMessage(seq)
//		if err != nil {
//			return errors.Wrap(err, "failed to call the repository")
//		}
//
//		result = msg
//		return nil
//	}); err != nil {
//		return message.Message{}, errors.Wrap(err, "transaction failed")
//	}
//
//	return result, nil
//}
//
//func (r ReadReceiveLogRepository) GetSequences(ref refs.Message) ([]common.ReceiveLogSequence, error) {
//	var result []common.ReceiveLogSequence
//
//	if err := r.db.View(func(tx *bbolt.Tx) error {
//		r, err := r.factory(tx)
//		if err != nil {
//			return errors.Wrap(err, "could not call the factory")
//		}
//
//		seq, err := r.ReceiveLog.GetSequences(ref)
//		if err != nil {
//			return errors.Wrap(err, "failed to call the repository")
//		}
//
//		result = seq
//		return nil
//	}); err != nil {
//		return nil, errors.Wrap(err, "transaction failed")
//	}
//
//	return result, nil
//}
