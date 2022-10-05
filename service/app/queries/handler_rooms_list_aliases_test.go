package queries_test

import (
	"context"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/transport/mocks"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestRoomsListAliasesHandler(t *testing.T) {
	c, err := di.BuildTestQueries(t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(fixtures.TestContext(t), 5*time.Second)
	defer cancel()

	roomIdentityRef := fixtures.SomeRefIdentity()
	roomAddress := network.NewAddress(fixtures.SomeString())

	connection := mocks.NewConnectionMock(ctx)
	connection.Mock(
		func(req *rpc.Request) []rpc.ResponseWithError {
			require.Equal(t, messages.RoomListAliasesProcedure.Typ(), req.Type())
			require.Equal(t, messages.RoomListAliasesProcedure.Name(), req.Name())
			require.Contains(t, string(req.Arguments()), c.LocalIdentity.String())

			return []rpc.ResponseWithError{
				{
					Value: rpc.NewResponse([]byte(`["alias1", "alias2"]`)),
					Err:   nil,
				},
				{
					Value: nil,
					Err:   rpc.ErrRemoteEnd,
				},
			}
		},
	)

	c.Dialer.MockPeer(roomIdentityRef.Identity(), roomAddress, connection)

	query, err := queries.NewRoomsListAliases(
		roomIdentityRef,
		roomAddress,
	)
	require.NoError(t, err)

	aliases, err := c.Queries.RoomsListAliases.Handle(ctx, query)
	require.NoError(t, err)

	require.Equal(t,
		[]aliases.Alias{
			aliases.MustNewAlias("alias1"),
			aliases.MustNewAlias("alias2"),
		},
		aliases,
	)
}
