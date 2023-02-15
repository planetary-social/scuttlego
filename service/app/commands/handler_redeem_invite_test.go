package commands_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/invites"
	"github.com/stretchr/testify/require"
)

func TestRedeemInviteHandler(t *testing.T) {
	tc, err := di.BuildTestCommands(t)
	require.NoError(t, err)

	invite := fixtures.SomeInvite()
	ctx := fixtures.TestContext(t)

	cmd, err := commands.NewRedeemInvite(invite)
	require.NoError(t, err)

	t.Run("no_error", func(t *testing.T) {
		tc.InviteRedeemer.RedeemInviteErr = nil
		err = tc.RedeemInvite.Handle(ctx, cmd)
		require.NoError(t, err)
	})

	t.Run("error_is_returned", func(t *testing.T) {
		tc.InviteRedeemer.RedeemInviteErr = fixtures.SomeError()
		err = tc.RedeemInvite.Handle(ctx, cmd)
		require.ErrorIs(t, err, tc.InviteRedeemer.RedeemInviteErr)
	})

	t.Run("known_error_results_in_nil", func(t *testing.T) {
		tc.InviteRedeemer.RedeemInviteErr = invites.ErrAlreadyFollowing
		err = tc.RedeemInvite.Handle(ctx, cmd)
		require.NoError(t, err)
	})
}
