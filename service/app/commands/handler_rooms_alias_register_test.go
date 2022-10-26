package commands_test

import (
	"context"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestRoomsAliasRegisterHandler_RemoteReturnsSomeData(t *testing.T) {
	c, err := di.BuildTestCommands(t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(fixtures.TestContext(t), 5*time.Second)
	defer cancel()

	roomIdentityRef := fixtures.SomeRefIdentity()
	roomAddress := network.NewAddress(fixtures.SomeString())

	alias := fixtures.SomeAlias()
	expectedAliasString := "https://somealias.example.com"

	connection := mocks.NewConnectionMock(ctx)
	connection.Mock(
		func(req *rpc.Request) []rpc.ResponseWithError {
			require.Equal(t, messages.RoomRegisterAliasProcedure.Typ(), req.Type())
			require.Equal(t, messages.RoomRegisterAliasProcedure.Name(), req.Name())
			require.Contains(t, string(req.Arguments()), alias.String())

			return []rpc.ResponseWithError{
				{
					Value: rpc.NewResponse([]byte(expectedAliasString)),
					Err:   nil,
				},
			}
		},
	)

	c.Dialer.MockPeer(roomIdentityRef.Identity(), roomAddress, connection)

	cmd, err := commands.NewRoomsAliasRegister(
		roomIdentityRef,
		roomAddress,
		alias,
	)
	require.NoError(t, err)

	aliasURL, err := c.RoomsAliasRegister.Handle(ctx, cmd)
	require.NoError(t, err)
	require.Equal(t, expectedAliasString, aliasURL.String())
}

func TestRoomsAliasRegisterHandler_RemoteTerminatesWithAnError(t *testing.T) {
	c, err := di.BuildTestCommands(t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(fixtures.TestContext(t), 5*time.Second)
	defer cancel()

	roomIdentityRef := fixtures.SomeRefIdentity()
	roomAddress := network.NewAddress(fixtures.SomeString())

	alias := fixtures.SomeAlias()

	connection := mocks.NewConnectionMock(ctx)
	connection.Mock(
		func(req *rpc.Request) []rpc.ResponseWithError {
			require.Equal(t, messages.RoomRegisterAliasProcedure.Typ(), req.Type())
			require.Equal(t, messages.RoomRegisterAliasProcedure.Name(), req.Name())
			require.Contains(t, string(req.Arguments()), alias.String())

			return []rpc.ResponseWithError{
				{
					Value: nil,
					Err:   rpc.ErrRemoteEnd,
				},
			}
		},
	)

	c.Dialer.MockPeer(roomIdentityRef.Identity(), roomAddress, connection)

	cmd, err := commands.NewRoomsAliasRegister(
		roomIdentityRef,
		roomAddress,
		alias,
	)
	require.NoError(t, err)

	_, err = c.RoomsAliasRegister.Handle(ctx, cmd)
	require.EqualError(t, err, "could not contact the pub and redeem the invite: received an error: remote end")
}
