package rpc_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/internal/mocks"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/messages"
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
	require.NoError(t, err)

	errCh := make(chan error)
	go func() {
		errCh <- handler.Handle(ctx, s, req)
	}()

	require.Eventually(t,
		func() bool {
			calls := commandHandler.HandleCalls()

			if len(calls) != 1 {
				return false
			}
			if !portal.Equal(calls[0].Portal()) {
				return false
			}
			if !target.Equal(calls[0].Target()) {
				return false
			}
			if !origin.Equal(calls[0].Origin()) {
				return false
			}
			if calls[0].Rwc() == nil {
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
	handleCalls []commands.AcceptTunnelConnect
	lock        sync.Mutex
}

func newAcceptTunnelConnectCommandHandlerMock() *acceptTunnelConnectCommandHandlerMock {
	return &acceptTunnelConnectCommandHandlerMock{}
}

func (a *acceptTunnelConnectCommandHandlerMock) Handle(ctx context.Context, cmd commands.AcceptTunnelConnect) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.handleCalls = append(a.handleCalls, cmd)
	return nil
}

func (a *acceptTunnelConnectCommandHandlerMock) HandleCalls() []commands.AcceptTunnelConnect {
	a.lock.Lock()
	defer a.lock.Unlock()
	tmp := make([]commands.AcceptTunnelConnect, len(a.handleCalls))
	copy(tmp, a.handleCalls)
	return tmp
}
