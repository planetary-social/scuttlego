package queries

import "github.com/boreq/errors"

type StatsResult struct {
	NumberOfMessages int
}

type MessageRepository interface {
	Count() (int, error)
}

type StatsHandler struct {
	repository MessageRepository
}

func NewStatsHandler(repository MessageRepository) *StatsHandler {
	return &StatsHandler{repository: repository}
}

func (h StatsHandler) Handle() (StatsResult, error) {
	numberOfMessages, err := h.repository.Count()
	if err != nil {
		return StatsResult{}, errors.Wrap(err, "could not get the number of messages")
	}

	return StatsResult{
		NumberOfMessages: numberOfMessages,
	}, nil
}
