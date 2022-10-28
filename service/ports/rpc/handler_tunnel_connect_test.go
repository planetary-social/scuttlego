package rpc_test

import (
	"context"
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux/mocks"
	"github.com/planetary-social/scuttlego/service/ports/rpc"
	"github.com/stretchr/testify/require"
)

func TestHandlerTunnelConnect_CallsCommandHandler(t *testing.T) {
	commandHandler := newAcceptTunnelConnectCommandHandlerMock()
	handler := rpc.NewHandlerTunnelConnect(commandHandler)

	require.Equal(t, messages.TunnelConnectProcedure, handler.Procedure())

	ctx := fixtures.TestContext(t)
	s := mocks.NewMockCloserStream()

	portal := fixtures.SomeRefIdentity()
	target := fixtures.SomeRefIdentity()
	origin := fixtures.SomeRefIdentity()

	args, err := messages.NewTunnelConnectToTargetArguments(portal, target, origin)
	require.NoError(t, err)

	req, err := messages.NewTunnelConnectToTarget(args)

	err = handler.Handle(ctx, s, req)
	require.NoError(t, err)

	require.Len(t, commandHandler.HandleCalls, 1)
	require.Equal(t, portal, commandHandler.HandleCalls[0].Portal())
	require.Equal(t, target, commandHandler.HandleCalls[0].Target())
	require.Equal(t, origin, commandHandler.HandleCalls[0].Origin())
	require.NotNil(t, commandHandler.HandleCalls[0].Rwc())
}

type acceptTunnelConnectCommandHandlerMock struct {
	HandleCalls []commands.AcceptTunnelConnect
}

func newAcceptTunnelConnectCommandHandlerMock() *acceptTunnelConnectCommandHandlerMock {
	return &acceptTunnelConnectCommandHandlerMock{}
}

func (a *acceptTunnelConnectCommandHandlerMock) Handle(ctx context.Context, cmd commands.AcceptTunnelConnect) error {
	a.HandleCalls = append(a.HandleCalls, cmd)
	return nil
}
