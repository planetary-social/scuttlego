package commands

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type Follow struct {
	target refs.Identity
}

func NewFollow(target refs.Identity) (Follow, error) {
	if target.IsZero() {
		return Follow{}, errors.New("zero value of target")
	}
	return Follow{target: target}, nil
}

func (f Follow) Target() refs.Identity {
	return f.target
}

func (f Follow) IsZero() bool {
	return f.target.IsZero()
}

type FollowHandler struct {
	transaction TransactionProvider
	local       identity.Private
	marshaler   content.Marshaler
	logger      logging.Logger
}

func NewFollowHandler(
	transaction TransactionProvider,
	local identity.Private,
	marshaler content.Marshaler,
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
	if cmd.IsZero() {
		return errors.New("zero value of cmd")
	}

	contactActions, err := known.NewContactActions([]known.ContactAction{known.ContactActionFollow})
	if err != nil {
		return errors.Wrap(err, "failed to create contact actions")
	}

	contact, err := known.NewContact(cmd.Target(), contactActions)
	if err != nil {
		return errors.Wrap(err, "failed to create a contact message")
	}

	rawContent, err := h.marshaler.Marshal(contact)
	if err != nil {
		return errors.Wrap(err, "failed to create message content")
	}

	myRef, err := refs.NewIdentityFromPublic(h.local.Public())
	if err != nil {
		return errors.Wrap(err, "could not create my own ref")
	}

	return h.transaction.Transact(func(adapters Adapters) error {
		return adapters.Feed.UpdateFeed(myRef.MainFeed(), func(feed *feeds.Feed) error {
			if _, err := feed.CreateMessage(rawContent, time.Now(), h.local); err != nil {
				return errors.Wrap(err, "failed to create a message")
			}
			return nil
		})
	})
}
