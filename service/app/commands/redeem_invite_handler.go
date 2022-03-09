package commands

import (
	"context"
	"time"

	"github.com/planetary-social/go-ssb/service/domain/feeds"
	"github.com/planetary-social/go-ssb/service/domain/feeds/content"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/invites"
	network2 "github.com/planetary-social/go-ssb/service/domain/network"
	boxstream2 "github.com/planetary-social/go-ssb/service/domain/network/boxstream"
	rpc2 "github.com/planetary-social/go-ssb/service/domain/network/rpc"
	"github.com/planetary-social/go-ssb/service/domain/network/rpc/messages"
	"github.com/planetary-social/go-ssb/service/domain/refs"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
)

type Dialer interface {
	DialWithInitializer(initializer network2.ClientPeerInitializer, remote identity.Public, addr network2.Address) (network2.Peer, error)
	Dial(remote identity.Public, address network2.Address) (network2.Peer, error)
}

type RedeemInvite struct {
	Invite invites.Invite
}

type RedeemInviteHandler struct {
	dialer         Dialer
	transaction    TransactionProvider
	networkKey     boxstream2.NetworkKey
	local          identity.Private
	requestHandler rpc2.RequestHandler
	logger         logging.Logger
}

func NewRedeemInviteHandler(
	dialer Dialer,
	transaction TransactionProvider,
	networkKey boxstream2.NetworkKey,
	local identity.Private,
	requestHandler rpc2.RequestHandler,
	logger logging.Logger,
) *RedeemInviteHandler {
	return &RedeemInviteHandler{
		dialer:         dialer,
		transaction:    transaction,
		networkKey:     networkKey,
		local:          local,
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
	follow, err := content.NewContact(cmd.Invite.Remote(), content.ContactActionFollow)
	if err != nil {
		return errors.Wrap(err, "could not create a follow message")
	}

	// todo indempotency

	myRef, err := refs.NewIdentityFromPublic(h.local.Public())
	if err != nil {
		return errors.Wrap(err, "could not create a local identity ref")
	}

	if err := h.transaction.Transact(func(adapters Adapters) error {
		if err := adapters.Feed.UpdateFeed(myRef.MainFeed(), func(feed *feeds.Feed) (*feeds.Feed, error) {
			if err := feed.CreateMessage(follow, time.Now(), h.local); err != nil {
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

func (h *RedeemInviteHandler) dial(cmd RedeemInvite) (network2.Peer, error) {
	local, err := identity.NewPrivateFromSeed(cmd.Invite.SecretKeySeed())
	if err != nil {
		return network2.Peer{}, errors.Wrap(err, "could not create a private identity")
	}

	handshaker, err := boxstream2.NewHandshaker(local, h.networkKey)
	if err != nil {
		return network2.Peer{}, errors.Wrap(err, "could not create a handshaker")
	}

	initializer := network2.NewPeerInitializer(handshaker, h.requestHandler, h.logger)

	peer, err := h.dialer.DialWithInitializer(initializer, cmd.Invite.Remote().Identity(), cmd.Invite.Address())
	if err != nil {
		return network2.Peer{}, errors.Wrap(err, "failed to dial")
	}

	return peer, nil
}

func (h *RedeemInviteHandler) createRequest() (*rpc2.Request, error) {
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
