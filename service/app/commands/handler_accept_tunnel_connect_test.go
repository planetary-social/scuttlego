package commands_test

import (
	"context"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestAcceptNewPeerHandler(t *testing.T) {
	tc, err := di.BuildTestCommands(t)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(fixtures.TestContext(t))

	origin := fixtures.SomeRefIdentity()
	portal := fixtures.SomeRefIdentity()
	rwc := newReadWriteCloserMock()

	target, err := refs.NewIdentityFromPublic(tc.Local)
	require.NoError(t, err)

	cmd, err := commands.NewAcceptTunnelConnect(origin, target, portal, rwc)
	require.NoError(t, err)

	errCh := make(chan error)
	go func() {
		errCh <- tc.AcceptTunnelConnect.Handle(ctx, cmd)
	}()

	require.Eventually(t,
		func() bool {
			return tc.PeerInitializer.InitializeServerPeerCalls() == 1
		},
		1*time.Second, 10*time.Millisecond,
	)

	require.Eventually(t,
		func() bool {
			return tc.NewPeerHandler.HandleNewPeerCalls() == 1
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

type readWriteCloserMock struct {
}

func newReadWriteCloserMock() *readWriteCloserMock {
	return &readWriteCloserMock{}
}

func (r readWriteCloserMock) Read(p []byte) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (r readWriteCloserMock) Write(p []byte) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (r readWriteCloserMock) Close() error {
	//TODO implement me
	panic("implement me")
}
