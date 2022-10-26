package invites_test

import (
	"context"
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/invites"
	"github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestInviteRedeemer_RedeemInviteReturnsNoErrorIfPubReturnsAResponse(t *testing.T) {
	logger := fixtures.TestLogger(t)
	dialer := newInviteDialerMock()
	redeemer := invites.NewInviteRedeemer(dialer, logger)

	ctx := fixtures.TestContext(t)
	invite := fixtures.SomeInvite()
	target := fixtures.SomePublicIdentity()
	conn := mocks.NewConnectionMock(ctx)
	conn.Mock(func(req *rpc.Request) []rpc.ResponseWithError {
		return []rpc.ResponseWithError{
			{
				Value: rpc.NewResponse(fixtures.SomeBytes()),
				Err:   nil,
			},
		}
	})
	dialer.DialReturnValue = transport.MustNewPeer(invite.Remote().Identity(), conn)

	err := redeemer.RedeemInvite(ctx, invite, target)
	require.NoError(t, err)
	require.Equal(t,
		[]inviteDialerMockDialCall{
			{
				Local:   identity.MustNewPrivateFromSeed(invite.SecretKeySeed()),
				Remote:  invite.Remote().Identity(),
				Address: invite.Address(),
			},
		},
		dialer.DialCalls,
	)
}

func TestInviteRedeemer_RedeemInviteReturnsAPredefinedErrorWhenThePeerIsAlreadyBeingFollowed(t *testing.T) {
	logger := fixtures.TestLogger(t)
	dialer := newInviteDialerMock()
	redeemer := invites.NewInviteRedeemer(dialer, logger)

	ctx := fixtures.TestContext(t)
	invite := fixtures.SomeInvite()
	target := fixtures.SomePublicIdentity()
	conn := mocks.NewConnectionMock(ctx)
	conn.Mock(func(req *rpc.Request) []rpc.ResponseWithError {
		return []rpc.ResponseWithError{
			{
				Value: nil,
				Err:   rpc.NewRemoteError([]byte(`{"message":"already following","name":"Error","stack":"Error: already following..."}`)),
			},
		}
	})
	dialer.DialReturnValue = transport.MustNewPeer(invite.Remote().Identity(), conn)

	err := redeemer.RedeemInvite(ctx, invite, target)
	require.ErrorIs(t, err, invites.ErrAlreadyFollowing)
	require.Equal(t,
		[]inviteDialerMockDialCall{
			{
				Local:   identity.MustNewPrivateFromSeed(invite.SecretKeySeed()),
				Remote:  invite.Remote().Identity(),
				Address: invite.Address(),
			},
		},
		dialer.DialCalls,
	)
}

func TestInviteRedeemer_RedeemInviteReturnsAGenericErrorIfADifferentErrorIsReturned(t *testing.T) {
	logger := fixtures.TestLogger(t)
	dialer := newInviteDialerMock()
	redeemer := invites.NewInviteRedeemer(dialer, logger)

	ctx := fixtures.TestContext(t)
	invite := fixtures.SomeInvite()
	target := fixtures.SomePublicIdentity()
	conn := mocks.NewConnectionMock(ctx)
	conn.Mock(func(req *rpc.Request) []rpc.ResponseWithError {
		return []rpc.ResponseWithError{
			{
				Value: nil,
				Err:   rpc.NewRemoteError([]byte(`some error`)),
			},
		}
	})
	dialer.DialReturnValue = transport.MustNewPeer(invite.Remote().Identity(), conn)

	err := redeemer.RedeemInvite(ctx, invite, target)
	require.Error(t, err)
	require.NotErrorIs(t, err, invites.ErrAlreadyFollowing)
	require.Equal(t,
		[]inviteDialerMockDialCall{
			{
				Local:   identity.MustNewPrivateFromSeed(invite.SecretKeySeed()),
				Remote:  invite.Remote().Identity(),
				Address: invite.Address(),
			},
		},
		dialer.DialCalls,
	)
}

type inviteDialerMock struct {
	DialReturnValue transport.Peer
	DialCalls       []inviteDialerMockDialCall
}

func newInviteDialerMock() *inviteDialerMock {
	return &inviteDialerMock{}
}

func (i *inviteDialerMock) Dial(ctx context.Context, local identity.Private, remote identity.Public, address network.Address) (transport.Peer, error) {
	i.DialCalls = append(i.DialCalls, inviteDialerMockDialCall{
		Local:   local,
		Remote:  remote,
		Address: address,
	})
	return i.DialReturnValue, nil
}

type inviteDialerMockDialCall struct {
	Local   identity.Private
	Remote  identity.Public
	Address network.Address
}
