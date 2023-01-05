package notx

import (
	"github.com/boreq/errors"
)

type NoTxMessageRepository struct {
	transaction TransactionProvider
}

func NewNoTxMessageRepository(transaction TransactionProvider) *NoTxMessageRepository {
	return &NoTxMessageRepository{
		transaction: transaction,
	}
}

func (r NoTxMessageRepository) Count() (int, error) {
	var result int

	if err := r.transaction.View(func(adapters TxAdapters) error {
		tmp, err := adapters.MessageRepository.Count()
		if err != nil {
			return errors.Wrap(err, "failed calling the repo")
		}

		result = tmp
		return nil
	}); err != nil {
		return 0, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}
