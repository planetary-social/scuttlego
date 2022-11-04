package rpc_test

import (
	"context"
	"testing"
	"time"

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

	ctx, cancel := context.WithCancel(fixtures.TestContext(t))
	s := mocks.NewMockCloserStream()

	portal := fixtures.SomeRefIdentity()
	target := fixtures.SomeRefIdentity()
	origin := fixtures.SomeRefIdentity()

	args, err := messages.NewTunnelConnectToTargetArguments(portal, target, origin)
	require.NoError(t, err)

	req, err := messages.NewTunnelConnectToTarget(args)

	errCh := make(chan error)
	go func() {
		errCh <- handler.Handle(ctx, s, req)
	}()

	require.Eventually(t,
		func() bool {
			if len(commandHandler.HandleCalls) != 1 {
				return false
			}
			if !portal.Equal(commandHandler.HandleCalls[0].Portal()) {
				return false
			}
			if !target.Equal(commandHandler.HandleCalls[0].Target()) {
				return false
			}
			if !origin.Equal(commandHandler.HandleCalls[0].Origin()) {
				return false
			}
			if commandHandler.HandleCalls[0].Rwc() == nil {
				return false
			}
			return true
		},
		1*time.Second, 10*time.Millisecond,
	)

	select {
	case <-errCh:
		t.Fatal("handler returned")
	case <-time.After(1 * time.Second):
		t.Log("ok, the handler is blocking")
	}

	cancel()

	select {
	case err := <-errCh:
		require.EqualError(t, err, "context canceled")
	case <-time.After(1 * time.Second):
		t.Fatal("timeout, the handler is still blocking")
	}
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
