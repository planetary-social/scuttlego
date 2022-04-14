package queries

import "github.com/boreq/errors"

type StatsResult struct {
	NumberOfMessages int
	NumberOfFeeds    int
}

type MessageRepository interface {
	// Count returns the number of stored messages.
	Count() (int, error)
}

type StatsHandler struct {
	messageRepository MessageRepository
	feedRepository    FeedRepository
}

func NewStatsHandler(messageRepository MessageRepository, feedRepository FeedRepository) *StatsHandler {
	return &StatsHandler{
		messageRepository: messageRepository,
		feedRepository:    feedRepository,
	}
}

func (h StatsHandler) Handle() (StatsResult, error) {
	numberOfMessages, err := h.messageRepository.Count()
	if err != nil {
		return StatsResult{}, errors.Wrap(err, "could not get the number of messages")
	}

	numberOfFeeds, err := h.feedRepository.Count()
	if err != nil {
		return StatsResult{}, errors.Wrap(err, "could not get the number of feeds")
	}

	return StatsResult{
		NumberOfMessages: numberOfMessages,
		NumberOfFeeds:    numberOfFeeds,
	}, nil
}
