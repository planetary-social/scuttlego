package commands_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/stretchr/testify/require"
)

func TestNewHandleIncomingEbtReplicate(t *testing.T) {
	version := 3
	format := messages.EbtReplicateFormatClassic
	stream := newMockStream()

	cmd, err := commands.NewHandleIncomingEbtReplicate(version, format, stream)
	require.NoError(t, err)

	require.Equal(t, version, cmd.Version())
	require.Equal(t, format, cmd.Format())
	require.Equal(t, stream, cmd.Stream())
}

type mockStream struct {
}

func newMockStream() *mockStream {
	return &mockStream{}
}
