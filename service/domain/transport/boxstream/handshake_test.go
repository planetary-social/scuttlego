package boxstream_test

import (
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/domain/transport/boxstream"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
	"time"
)

func TestHandshaker(t *testing.T) {
	networkKey := boxstream.NewDefaultNetworkKey()

	peer1 := fixtures.SomePrivateIdentity()
	handshaker1, err := boxstream.NewHandshaker(peer1, networkKey)
	require.NoError(t, err)

	peer2 := fixtures.SomePrivateIdentity()
	handshaker2, err := boxstream.NewHandshaker(peer2, networkKey)
	require.NoError(t, err)

	oneToTwoReader, oneToTwoWriter := io.Pipe()
	twoToOneReader, twoToOneWriter := io.Pipe()

	conn1 := newMockConnection(twoToOneReader, oneToTwoWriter)
	defer conn1.Close()

	conn2 := newMockConnection(oneToTwoReader, twoToOneWriter)
	defer conn2.Close()

	errCh := make(chan error)

	go func() {
		_, err := handshaker1.OpenClientStream(conn1, peer2.Public())
		errCh <- err
	}()

	go func() {
		_, err := handshaker2.OpenServerStream(conn2)
		errCh <- err
	}()

	for i := 0; i < 2; i++ {
		select {
		case err := <-errCh:
			require.NoError(t, err)
		case <-time.After(1 * time.Second):
			t.Fatal("timeout")
		}
	}
}

type mockConnection struct {
	read  io.ReadCloser
	write io.WriteCloser
}

func newMockConnection(read io.ReadCloser, write io.WriteCloser) *mockConnection {
	return &mockConnection{
		read:  read,
		write: write,
	}
}

func (m mockConnection) Read(p []byte) (n int, err error) {
	return m.read.Read(p)
}

func (m mockConnection) Write(p []byte) (n int, err error) {
	return m.write.Write(p)
}

func (m mockConnection) Close() error {
	var err error
	err = multierror.Append(err, m.read.Close())
	err = multierror.Append(err, m.write.Close())
	return err
}
