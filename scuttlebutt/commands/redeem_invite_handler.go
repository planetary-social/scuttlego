package commands

import (
	"context"
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/identity"
	"github.com/planetary-social/go-ssb/invites"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/network"
	"github.com/planetary-social/go-ssb/network/boxstream"
	"github.com/planetary-social/go-ssb/network/rpc"
	"github.com/planetary-social/go-ssb/network/rpc/messages"
	"github.com/planetary-social/go-ssb/refs"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/content"
	"time"
)

type Dialer interface {
	DialWithInitializer(initializer network.ClientPeerInitializer, remote identity.Public, addr network.Address) (network.Peer, error)
}

type RedeemInvite struct {
	Invite invites.Invite
}

type RedeemInviteHandler struct {
	dialer         Dialer
	networkKey     boxstream.NetworkKey
	local          identity.Private
	storage        FeedStorage
	requestHandler rpc.RequestHandler
	logger         logging.Logger
}

func NewRedeemInviteHandler(
	dialer Dialer,
	networkKey boxstream.NetworkKey,
	local identity.Private,
	storage FeedStorage,
	requestHandler rpc.RequestHandler,
	logger logging.Logger,
) *RedeemInviteHandler {
	return &RedeemInviteHandler{
		dialer:         dialer,
		networkKey:     networkKey,
		local:          local,
		storage:        storage,
		requestHandler: requestHandler,
		logger:         logger.New("follow_handler"),
	}
}

func (h *RedeemInviteHandler) Handle(ctx context.Context, cmd RedeemInvite) error {
	if err := h.redeemInvite(ctx, cmd); err != nil {
		return errors.Wrap(err, "could not contact the pub and redeem the invite")
	}

	// todo check reply

	// todo publish contact and pub content

	// todo main feed or should the invite contain a feed ref?
	follow, err := content.NewContact(cmd.Invite.Remote().MainFeed(), content.ContactActionFollow)
	if err != nil {
		return errors.Wrap(err, "could not create a follow message")
	}

	// todo indempotency

	myRef, err := refs.NewIdentityFromPublic(h.local.Public())
	if err != nil {
		return errors.Wrap(err, "could not create a local identity ref")
	}

	if err := h.storage.UpdateFeed(myRef.MainFeed(), func(feed *feeds.Feed) (*feeds.Feed, error) {
		if feed == nil {
			// todo just create feed if it doesn't exist
			feed, err = feeds.NewFeedFromMessageContent(follow, time.Now(), h.local)
			if err != nil {
				return nil, errors.Wrap(err, "could not create a new feed")
			}

			return feed, nil
		}

		if err := feed.CreateMessage(follow, time.Now(), h.local); err != nil {
			return nil, errors.Wrap(err, "could not append a message")
		}

		return feed, nil
	}); err != nil {
		return errors.Wrap(err, "failed to update the feed")
	}

	return errors.New("not implemented")
}

func (h *RedeemInviteHandler) redeemInvite(ctx context.Context, cmd RedeemInvite) error {
	peer, err := h.dial(cmd)
	if err != nil {
		return errors.Wrap(err, "could not dial the pub")
	}

	req, err := h.createRequest()
	if err != nil {
		return errors.Wrap(err, "could not create a request")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rs, err := peer.Conn().PerformRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to perform a request")
	}
	defer rs.Close()

	response, ok := <-rs.Channel()
	if !ok {
		return errors.New("channel closed")
	}

	if response.Err != nil {
		return errors.Wrap(err, "received an error")
	}

	h.logger.WithField("response", string(response.Value.Bytes())).Debug("response received")

	return nil
}

func (h *RedeemInviteHandler) dial(cmd RedeemInvite) (network.Peer, error) {
	local, err := identity.NewPrivateFromSeed(cmd.Invite.SecretKeySeed())
	if err != nil {
		return network.Peer{}, errors.Wrap(err, "could not create a private identity")
	}

	handshaker, err := boxstream.NewHandshaker(local, h.networkKey)
	if err != nil {
		return network.Peer{}, errors.Wrap(err, "could not create a handshaker")
	}

	initializer := network.NewPeerInitializer(handshaker, h.requestHandler, h.logger)

	peer, err := h.dialer.DialWithInitializer(initializer, cmd.Invite.Remote().Identity(), cmd.Invite.Address())
	if err != nil {
		return network.Peer{}, errors.Wrap(err, "failed to dial")
	}

	return peer, nil
}

func (h *RedeemInviteHandler) createRequest() (*rpc.Request, error) {
	public, err := refs.NewIdentityFromPublic(h.local.Public())
	if err != nil {
		return nil, errors.Wrap(err, "could not create a ref")
	}

	args, err := messages.NewInviteUseArguments(public)
	if err != nil {
		return nil, errors.Wrap(err, "could not create args")
	}

	req, err := messages.NewInviteUse(args)
	if err != nil {
		return nil, errors.Wrap(err, "could not create args")
	}

	return req, nil
}
