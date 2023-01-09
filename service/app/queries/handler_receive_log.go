package queries

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type ReceiveLogRepository interface {
	// List returns messages from the log starting with the provided sequence.
	// This is supposed to simulate the behaviour of go-ssb's receive log as
	// such a concept doesn't exist within this implementation. The log is zero
	// indexed. If limit isn't positive an error is returned. Sequence has
	// nothing to do with the sequence field of Scuttlebutt messages.
	List(startSeq common.ReceiveLogSequence, limit int) ([]LogMessage, error)

	// GetMessage returns the message that the provided receive log sequence
	// points to.
	GetMessage(seq common.ReceiveLogSequence) (message.Message, error)

	// GetSequences returns the sequences assigned to a message in the receive
	// log. If an error isn't returned then the slice will have at least one
	// element.
	GetSequences(ref refs.Message) ([]common.ReceiveLogSequence, error)
}

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
	repository ReceiveLogRepository
}

func NewReceiveLogHandler(repository ReceiveLogRepository) *ReceiveLogHandler {
	return &ReceiveLogHandler{repository: repository}
}

func (h *ReceiveLogHandler) Handle(query ReceiveLog) ([]LogMessage, error) {
	if query.IsZero() {
		return nil, errors.New("zero value of query")
	}

	return h.repository.List(query.StartSeq(), query.Limit())
}
