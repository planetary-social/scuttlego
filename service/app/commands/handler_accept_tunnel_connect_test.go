package commands_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestAcceptNewPeerHandler(t *testing.T) {
	tc, err := di.BuildTestCommands(t)
	require.NoError(t, err)

	ctx := fixtures.TestContext(t)

	origin := fixtures.SomeRefIdentity()
	portal := fixtures.SomeRefIdentity()
	rwc := newReadWriteCloserMock()

	target, err := refs.NewIdentityFromPublic(tc.Local)
	require.NoError(t, err)

	cmd, err := commands.NewAcceptTunnelConnect(origin, target, portal, rwc)
	require.NoError(t, err)

	err = tc.AcceptTunnelConnect.Handle(ctx, cmd)
	require.NoError(t, err)

	require.Equal(t, 1, tc.PeerInitializer.InitializeServerPeerCalls())
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
