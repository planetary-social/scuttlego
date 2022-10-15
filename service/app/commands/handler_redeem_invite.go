package commands

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/invites"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type RedeemInvite struct {
	Invite invites.Invite
}

type RedeemInviteHandler struct {
	dialer                Dialer
	transaction           TransactionProvider
	networkKey            boxstream.NetworkKey
	local                 identity.Private
	requestHandler        rpc.RequestHandler
	marshaler             formats.Marshaler
	connectionIdGenerator *rpc.ConnectionIdGenerator
	currentTimeProvider   CurrentTimeProvider
	logger                logging.Logger
}

func NewRedeemInviteHandler(
	dialer Dialer,
	transaction TransactionProvider,
	networkKey boxstream.NetworkKey,
	local identity.Private,
	requestHandler rpc.RequestHandler,
	marshaler formats.Marshaler,
	connectionIdGenerator *rpc.ConnectionIdGenerator,
	currentTimeProvider CurrentTimeProvider,
	logger logging.Logger,
) *RedeemInviteHandler {
	return &RedeemInviteHandler{
		dialer:                dialer,
		transaction:           transaction,
		networkKey:            networkKey,
		local:                 local,
		requestHandler:        requestHandler,
		marshaler:             marshaler,
		connectionIdGenerator: connectionIdGenerator,
		currentTimeProvider:   currentTimeProvider,
		logger:                logger.New("follow_handler"),
	}
}

func (h *RedeemInviteHandler) Handle(ctx context.Context, cmd RedeemInvite) error {
	if err := h.redeemInvite(ctx, cmd); err != nil {
		return errors.Wrap(err, "could not contact the pub and redeem the invite")
	}

	// todo check reply

	// todo publish contact and pub message?

	// todo main feed or should the invite contain a feed ref?
	contactActions, err := content.NewContactActions([]content.ContactAction{content.ContactActionFollow})
	if err != nil {
		return errors.Wrap(err, "failed to create contact actions")
	}

	follow, err := content.NewContact(cmd.Invite.Remote(), contactActions)
	if err != nil {
		return errors.Wrap(err, "could not create a follow message")
	}

	followContent, err := h.marshaler.Marshal(follow)
	if err != nil {
		return errors.Wrap(err, "failed to create message content")
	}

	// todo indempotency

	myRef, err := refs.NewIdentityFromPublic(h.local.Public())
	if err != nil {
		return errors.Wrap(err, "could not create a local identity ref")
	}

	if err := h.transaction.Transact(func(adapters Adapters) error {
		if err := adapters.Feed.UpdateFeed(myRef.MainFeed(), func(feed *feeds.Feed) error {
			if _, err := feed.CreateMessage(followContent, time.Now(), h.local); err != nil {
				return errors.Wrap(err, "could not append a message")
			}
			return nil
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
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	peer, err := h.dial(ctx, cmd)
	if err != nil {
		return errors.Wrap(err, "could not dial the pub")
	}

	req, err := h.createRequest()
	if err != nil {
		return errors.Wrap(err, "could not create a request")
	}

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

func (h *RedeemInviteHandler) dial(ctx context.Context, cmd RedeemInvite) (transport.Peer, error) {
	local, err := identity.NewPrivateFromSeed(cmd.Invite.SecretKeySeed())
	if err != nil {
		return transport.Peer{}, errors.Wrap(err, "could not create a private identity")
	}

	handshaker, err := boxstream.NewHandshaker(local, h.networkKey, h.currentTimeProvider)
	if err != nil {
		return transport.Peer{}, errors.Wrap(err, "could not create a handshaker")
	}

	initializer := transport.NewPeerInitializer(handshaker, h.requestHandler, h.connectionIdGenerator, h.logger)

	peer, err := h.dialer.DialWithInitializer(ctx, initializer, cmd.Invite.Remote().Identity(), cmd.Invite.Address())
	if err != nil {
		return transport.Peer{}, errors.Wrap(err, "failed to dial")
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
