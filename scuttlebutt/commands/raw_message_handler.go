package commands

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/refs"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
)

type FeedStorage interface {
	// UpdateFeed updates the specified feed by calling the provided function on it. Presumably the modification
	// should happen in a transaction. If the feed doesn't exist the function receives a nil pointer. In that
	// case a new feed can be created and returned.
	UpdateFeed(ref refs.Feed, f func(feed *feeds.Feed) (*feeds.Feed, error)) error
}

type RawMessageIdentifier interface {
	IdentifyRawMessage(raw message.RawMessage) (message.Message, error)
}

// RawMessageHandler processes incoming content retrieved through various replication methods.
type RawMessageHandler struct {
	storage    FeedStorage
	identifier RawMessageIdentifier
	logger     logging.Logger
}

func NewRawMessageHandler(storage FeedStorage, identifier RawMessageIdentifier, logger logging.Logger) *RawMessageHandler {
	return &RawMessageHandler{
		storage:    storage,
		identifier: identifier,
		logger:     logger.New("raw_message_handler"),
	}
}

func (h *RawMessageHandler) Handle(rawMsg message.RawMessage) error {
	h.logger.WithField("raw", string(rawMsg.Bytes())).Debug("handling a raw message")

	msg, err := h.identifier.IdentifyRawMessage(rawMsg)
	if err != nil {
		return errors.Wrap(err, "failed to identify the raw message")
	}

	if err := h.storage.UpdateFeed(msg.Feed(), func(feed *feeds.Feed) (*feeds.Feed, error) {
		if feed == nil {
			feed, err = feeds.NewFeed(msg)
			if err != nil {
				return nil, errors.Wrap(err, "could not create a new feed")
			}

			return feed, nil
		}

		if err := feed.AppendMessage(msg); err != nil {
			return nil, errors.Wrap(err, "could not append a message")
		}

		return feed, nil
	}); err != nil {
		return errors.Wrap(err, "failed to update the feed")
	}

	return nil
}
