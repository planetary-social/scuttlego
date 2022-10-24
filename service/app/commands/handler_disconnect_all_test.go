package commands_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/stretchr/testify/require"
)

func TestDisconnectAllHandler(t *testing.T) {
	tc, err := di.BuildTestCommands(t)
	require.NoError(t, err)

	err = tc.DisconnectAll.Handle()
	require.NoError(t, err)

	require.Equal(t, 1, tc.PeerManager.DisconnectAllCalls())
}
