package queries

import (
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
)

type ReceiveLogRepository interface {
	// Get returns messages from the log starting with the provided sequence. The log is zero indexed. If limit isn't
	// positive an error is returned.
	Get(startSeq int, limit int) ([]message.Message, error)
}

type GetReceiveLog struct {
	StartSeq int
	Limit    int
}

type GetReceiveLogHandler struct {
	repository ReceiveLogRepository
}

func NewGetReceiveLogHandler(repository ReceiveLogRepository) *GetReceiveLogHandler {
	return &GetReceiveLogHandler{repository: repository}
}

func (h *GetReceiveLogHandler) Handle(query GetReceiveLog) ([]message.Message, error) {
	return h.repository.Get(query.StartSeq, query.Limit)
}
