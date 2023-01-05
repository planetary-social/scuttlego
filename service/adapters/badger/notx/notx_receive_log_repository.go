package notx

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type NoTxReceiveLogRepository struct {
	transaction TransactionProvider
}

func NewNoTxReceiveLogRepository(transaction TransactionProvider) *NoTxReceiveLogRepository {
	return &NoTxReceiveLogRepository{
		transaction: transaction,
	}
}

func (r NoTxReceiveLogRepository) List(startSeq common.ReceiveLogSequence, limit int) ([]queries.LogMessage, error) {
	var result []queries.LogMessage

	if err := r.transaction.View(func(adapters TxAdapters) error {
		tmp, err := adapters.ReceiveLogRepository.List(startSeq, limit)
		if err != nil {
			return errors.Wrap(err, "failed to call the repository")
		}

		result = tmp
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}

func (r NoTxReceiveLogRepository) GetMessage(seq common.ReceiveLogSequence) (message.Message, error) {
	var result message.Message

	if err := r.transaction.View(func(adapters TxAdapters) error {
		tmp, err := adapters.ReceiveLogRepository.GetMessage(seq)
		if err != nil {
			return errors.Wrap(err, "failed to call the repository")
		}

		result = tmp
		return nil
	}); err != nil {
		return message.Message{}, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}

func (r NoTxReceiveLogRepository) GetSequences(ref refs.Message) ([]common.ReceiveLogSequence, error) {
	var result []common.ReceiveLogSequence

	if err := r.transaction.View(func(adapters TxAdapters) error {
		tmp, err := adapters.ReceiveLogRepository.GetSequences(ref)
		if err != nil {
			return errors.Wrap(err, "failed to call the repository")
		}

		result = tmp
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}
