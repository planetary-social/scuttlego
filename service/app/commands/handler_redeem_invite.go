package commands

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/invites"
)

type InviteRedeemer interface {
	RedeemInvite(ctx context.Context, invite invites.Invite, target identity.Public) error
}

type RedeemInvite struct {
	Invite invites.Invite
}

type RedeemInviteHandler struct {
	redeemer InviteRedeemer
	local    identity.Private
	logger   logging.Logger
}

func NewRedeemInviteHandler(
	redeemer InviteRedeemer,
	local identity.Private,
	logger logging.Logger,
) *RedeemInviteHandler {
	return &RedeemInviteHandler{
		redeemer: redeemer,
		local:    local,
		logger:   logger.New("redeem_invite"),
	}
}

func (h *RedeemInviteHandler) Handle(ctx context.Context, cmd RedeemInvite) error {
	if err := h.redeemer.RedeemInvite(ctx, cmd.Invite, h.local.Public()); err != nil {
		if errors.Is(err, invites.ErrAlreadyFollowing) {
			h.logger.Debug("pub reported that it is already following the user")
			return nil
		}
		return errors.Wrap(err, "could not contact the pub and redeem the invite")
	}
	return nil
}
