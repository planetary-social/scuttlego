package invites

import (
	"context"
	"strings"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type InviteDialer interface {
	Dial(ctx context.Context, local identity.Private, remote identity.Public, address network.Address) (transport.Peer, error)
}

var ErrAlreadyFollowing = errors.New("already following")

type InviteRedeemer struct {
	dialer InviteDialer
	logger logging.Logger
}

func NewInviteRedeemer(dialer InviteDialer, logger logging.Logger) *InviteRedeemer {
	return &InviteRedeemer{
		dialer: dialer,
		logger: logger.New("invite_redeemer"),
	}
}

func (h *InviteRedeemer) RedeemInvite(ctx context.Context, invite Invite, target identity.Public) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	peer, err := h.dial(ctx, invite)
	if err != nil {
		return errors.Wrap(err, "could not dial the pub")
	}

	req, err := h.createRequest(target)
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

	if err := response.Err; err != nil {
		var remoteErr rpc.RemoteError
		if errors.As(err, &remoteErr) {
			if strings.Contains(string(remoteErr.Response()), "already following") {
				return ErrAlreadyFollowing
			}
		}
		return errors.Wrap(err, "received an error")
	}

	h.logger.Debug().WithField("response", string(response.Value.Bytes())).Message("response received")

	return nil
}

func (h *InviteRedeemer) dial(ctx context.Context, invite Invite) (transport.Peer, error) {
	local, err := identity.NewPrivateFromSeed(invite.SecretKeySeed())
	if err != nil {
		return transport.Peer{}, errors.Wrap(err, "could not create a private identity")
	}

	peer, err := h.dialer.Dial(ctx, local, invite.Remote().Identity(), invite.Address())
	if err != nil {
		return transport.Peer{}, errors.Wrap(err, "failed to dial")
	}

	return peer, nil
}

func (h *InviteRedeemer) createRequest(target identity.Public) (*rpc.Request, error) {
	public, err := refs.NewIdentityFromPublic(target)
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
