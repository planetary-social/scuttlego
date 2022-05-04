package boxstream_test

import (
	"io"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/domain/transport/boxstream"
	"github.com/stretchr/testify/require"
)

func TestHandshaker(t *testing.T) {
	t.Parallel()
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

	var stream1 *boxstream.Stream
	var stream2 *boxstream.Stream

	go func() {
		s, err := handshaker1.OpenClientStream(conn1, peer2.Public())
		stream1 = s
		errCh <- err
	}()

	go func() {
		s, err := handshaker2.OpenServerStream(conn2)
		stream2 = s
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

	// test reading and writing to confirm that secrets are set correctly
	testWriteRead(t, stream1, stream2, []byte("test"))
	testWriteRead(t, stream2, stream1, []byte("test"))
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
