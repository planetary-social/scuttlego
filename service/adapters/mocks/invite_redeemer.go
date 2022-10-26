package mocks

import (
	"context"

	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/invites"
)

type InviteRedeemerMock struct {
	RedeemInviteErr error
}

func NewInviteRedeemerMock() *InviteRedeemerMock {
	return &InviteRedeemerMock{}
}

func (i InviteRedeemerMock) RedeemInvite(ctx context.Context, invite invites.Invite, target identity.Public) error {
	return i.RedeemInviteErr
}
