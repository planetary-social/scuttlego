package commands

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
)

type RawMessageIdentifier interface {
	IdentifyRawMessage(raw message.RawMessage) (message.Message, error)
}

// RawMessageHandler processes incoming content retrieved through various replication methods.
type RawMessageHandler struct {
	transaction TransactionProvider
	identifier  RawMessageIdentifier
	logger      logging.Logger
}

func NewRawMessageHandler(transaction TransactionProvider, identifier RawMessageIdentifier, logger logging.Logger) *RawMessageHandler {
	return &RawMessageHandler{
		transaction: transaction,
		identifier:  identifier,
		logger:      logger.New("raw_message_handler"),
	}
}

func (h *RawMessageHandler) Handle(rawMsg message.RawMessage) error {
	h.logger.WithField("raw", string(rawMsg.Bytes())).Debug("handling a raw message")

	msg, err := h.identifier.IdentifyRawMessage(rawMsg)
	if err != nil {
		return errors.Wrap(err, "failed to identify the raw message")
	}

	if err := h.transaction.Transact(func(adapters Adapters) error {
		has, err := adapters.SocialGraph.HasContact(msg.Author())
		if err != nil {
			return errors.Wrap(err, "could not check the social graph")
		}

		if !has {
			return nil // do nothing as this contact is not in our social graph
		}

		if err := adapters.Feed.UpdateFeed(msg.Feed(), func(feed *feeds.Feed) (*feeds.Feed, error) {
			if err := feed.AppendMessage(msg); err != nil {
				return nil, errors.Wrap(err, "could not append a message")
			}

			return feed, nil
		}); err != nil {
			return errors.Wrap(err, "failed to update the feed")
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
