package queries

import (
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
)

type ReceiveLogRepository interface {
	Next(lastSeq uint64) ([]message.Message, error)
}

type GetReceiveLog struct {
	LastSeq uint64
}

type GetReceiveLogHandler struct {
	repository ReceiveLogRepository
}

func NewGetReceiveLogHandler(repository ReceiveLogRepository) *GetReceiveLogHandler {
	return &GetReceiveLogHandler{repository: repository}
}

func (h *GetReceiveLogHandler) Handle(query GetReceiveLog) ([]message.Message, error) {
	return h.repository.Next(query.LastSeq)
}
