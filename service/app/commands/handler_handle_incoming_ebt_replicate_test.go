package commands_test

import (
	"context"
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
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

func (m mockStream) RemoteIdentity() identity.Public {
	return fixtures.SomePublicIdentity()
}

func (m mockStream) IncomingMessages(ctx context.Context) <-chan ebt.IncomingMessage {
	ch := make(chan ebt.IncomingMessage)
	close(ch)
	return ch
}

func (m mockStream) SendNotes(notes messages.EbtReplicateNotes) error {
	return errors.New("not implemented")
}

func (m mockStream) SendMessage(msg *message.Message) error {
	return errors.New("not implemented")
}
