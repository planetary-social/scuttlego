package queries

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
)

type ReceiveLogRepository interface {
	// Get returns messages from the log starting with the provided sequence. This is supposed to simulate
	// the behaviour of go-ssb's receive log as such a concept doesn't exist within this implementation. The log is
	// zero indexed. If limit isn't positive an error is returned. Sequence has nothing to do with the sequence
	// field of Scuttlebutt messages.
	Get(startSeq int, limit int) ([]message.Message, error)
}

type ReceiveLog struct {
	StartSeq int
	Limit    int
}

type ReceiveLogHandler struct {
	repository ReceiveLogRepository
}

func NewReceiveLogHandler(repository ReceiveLogRepository) *ReceiveLogHandler {
	return &ReceiveLogHandler{repository: repository}
}

func (h *ReceiveLogHandler) Handle(query ReceiveLog) ([]message.Message, error) {
	if query.StartSeq < 0 {
		return nil, errors.New("start seq can't be negative")
	}

	if query.Limit <= 0 {
		return nil, errors.New("limit must be positive")
	}

	return h.repository.Get(query.StartSeq, query.Limit)
}
