package queries

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/common"
)

type ReceiveLog struct {
	// Only messages with a sequence greater or equal to the start sequence are
	// returned.
	startSeq common.ReceiveLogSequence

	// Limit specifies the max number of messages which will be returned. Limit
	// must be positive.
	limit int
}

func NewReceiveLog(startSeq common.ReceiveLogSequence, limit int) (ReceiveLog, error) {
	if limit <= 0 {
		return ReceiveLog{}, errors.New("limit must be positive")
	}

	return ReceiveLog{startSeq: startSeq, limit: limit}, nil
}

func (r ReceiveLog) StartSeq() common.ReceiveLogSequence {
	return r.startSeq
}

func (r ReceiveLog) Limit() int {
	return r.limit
}

func (r ReceiveLog) IsZero() bool {
	return r == ReceiveLog{}
}

type ReceiveLogHandler struct {
	transaction TransactionProvider
}

func NewReceiveLogHandler(transaction TransactionProvider) *ReceiveLogHandler {
	return &ReceiveLogHandler{transaction: transaction}
}

func (h *ReceiveLogHandler) Handle(query ReceiveLog) ([]LogMessage, error) {
	if query.IsZero() {
		return nil, errors.New("zero value of query")
	}

	var result []LogMessage

	if err := h.transaction.Transact(func(adapters Adapters) error {
		tmp, err := adapters.ReceiveLog.List(query.StartSeq(), query.Limit())
		if err != nil {
			return errors.Wrap(err, "error listing messages")
		}
		result = tmp
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}
