package commands

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type PublishRaw struct {
	Content []byte
}

type PublishRawHandler struct {
	transaction TransactionProvider
	local       identity.Private
	logger      logging.Logger
}

func NewPublishRawHandler(transaction TransactionProvider, local identity.Private, logger logging.Logger) *PublishRawHandler {
	return &PublishRawHandler{
		transaction: transaction,
		local:       local,
		logger:      logger,
	}
}

func (h *PublishRawHandler) Handle(cmd PublishRaw) (refs.Message, error) {
	content, err := message.NewRawMessageContent(cmd.Content)
	if err != nil {
		return refs.Message{}, errors.Wrap(err, "could not create raw message content")
	}

	myRef, err := refs.NewIdentityFromPublic(h.local.Public())
	if err != nil {
		return refs.Message{}, errors.Wrap(err, "could not create my own ref")
	}

	var id refs.Message

	if err := h.transaction.Transact(func(adapters Adapters) error {
		return adapters.Feed.UpdateFeed(myRef.MainFeed(), func(feed *feeds.Feed) error {
			var err error
			id, err = feed.CreateMessage(content, time.Now(), h.local)
			if err != nil {
				return errors.Wrap(err, "failed to create a message")
			}
			return nil
		})
	}); err != nil {
		return refs.Message{}, errors.Wrap(err, "transaction failed")
	}

	return id, nil
}
