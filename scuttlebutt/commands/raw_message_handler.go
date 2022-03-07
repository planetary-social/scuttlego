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
	//h.logger.WithField("raw", string(rawMsg.Bytes())).Debug("handling a raw message")

	msg, err := h.identifier.IdentifyRawMessage(rawMsg)
	if err != nil {
		return errors.Wrap(err, "failed to identify the raw message")
	}

	h.logger.
		WithField("sequence", msg.Sequence()).
		WithField("feed", msg.Feed().String()).
		Debug("handling a new message")

	if err := h.transaction.Transact(func(adapters Adapters) error {
		return h.storeMessage(adapters, msg)
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}

func (h *RawMessageHandler) storeMessage(adapters Adapters, msg message.Message) error {
	socialGraph, err := adapters.SocialGraph.GetSocialGraph()
	if err != nil {
		return errors.Wrap(err, "could not load the social graph")
	}

	if !socialGraph.HasContact(msg.Author()) {
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
}
