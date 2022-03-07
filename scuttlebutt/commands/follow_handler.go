package commands

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/identity"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/refs"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/content"
)

type Follow struct {
	Target refs.Identity
}

type FollowHandler struct {
	transaction TransactionProvider
	local       identity.Private
	logger      logging.Logger
}

func NewFollowHandler(
	transaction TransactionProvider,
	local identity.Private,
	logger logging.Logger,
) *FollowHandler {
	return &FollowHandler{
		transaction: transaction,
		local:       local,
		logger:      logger.New("follow_handler"),
	}
}

func (h *FollowHandler) Handle(cmd Follow) error {
	contact, err := content.NewContact(cmd.Target, content.ContactActionFollow)
	if err != nil {
		return errors.Wrap(err, "failed to create a contact message")
	}

	myRef, err := refs.NewIdentityFromPublic(h.local.Public())
	if err != nil {
		return errors.Wrap(err, "could not create my own ref")
	}

	return h.transaction.Transact(func(adapters Adapters) error {
		return adapters.Feed.UpdateFeed(myRef.MainFeed(), func(feed *feeds.Feed) (*feeds.Feed, error) {
			if err := feed.CreateMessage(contact, time.Now(), h.local); err != nil {
				return nil, errors.Wrap(err, "failed to create a message")
			}
			return feed, nil
		})
	})
}
