package commands

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type Follow struct {
	Target refs.Identity
}

type FollowHandler struct {
	transaction TransactionProvider
	local       identity.Private
	marshaler   formats.Marshaler
	logger      logging.Logger
}

func NewFollowHandler(
	transaction TransactionProvider,
	local identity.Private,
	marshaler formats.Marshaler,
	logger logging.Logger,
) *FollowHandler {
	return &FollowHandler{
		transaction: transaction,
		local:       local,
		marshaler:   marshaler,
		logger:      logger.New("follow_handler"),
	}
}

func (h *FollowHandler) Handle(cmd Follow) error {
	contact, err := content.NewContact(cmd.Target, content.ContactActionFollow)
	if err != nil {
		return errors.Wrap(err, "failed to create a contact message")
	}

	content, err := h.marshaler.Marshal(contact)
	if err != nil {
		return errors.Wrap(err, "failed to create message content")
	}

	myRef, err := refs.NewIdentityFromPublic(h.local.Public())
	if err != nil {
		return errors.Wrap(err, "could not create my own ref")
	}

	return h.transaction.Transact(func(adapters Adapters) error {
		return adapters.Feed.UpdateFeed(myRef.MainFeed(), func(feed *feeds.Feed) (*feeds.Feed, error) {
			if _, err := feed.CreateMessage(content, time.Now(), h.local); err != nil {
				return nil, errors.Wrap(err, "failed to create a message")
			}
			return feed, nil
		})
	})
}
