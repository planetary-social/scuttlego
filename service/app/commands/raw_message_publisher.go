package commands

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type TransactionRawMessagePublisher struct {
	transaction TransactionProvider
}

func NewTransactionRawMessagePublisher(transaction TransactionProvider) *TransactionRawMessagePublisher {
	return &TransactionRawMessagePublisher{transaction: transaction}
}

func (h *TransactionRawMessagePublisher) Publish(identity identity.Private, content message.RawMessageContent) (refs.Message, error) {
	if identity.IsZero() {
		return refs.Message{}, errors.New("zero value of identity")
	}

	if content.IsZero() {
		return refs.Message{}, errors.New("zero value of content")
	}

	identityRef, err := refs.NewIdentityFromPublic(identity.Public())
	if err != nil {
		return refs.Message{}, errors.Wrap(err, "could not create the identity ref")
	}

	var id refs.Message

	if err := h.transaction.Transact(func(adapters Adapters) error {
		return adapters.Feed.UpdateFeed(identityRef.MainFeed(), func(feed *feeds.Feed) error {
			var err error
			id, err = feed.CreateMessage(content, time.Now(), identity)
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
