package boxstream_test

import (
	"io"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
	"github.com/stretchr/testify/require"
)

func TestHandshaker_ConnectionDoesNotImplementSetDeadliner(t *testing.T) {
	t.Parallel()

	networkKey := boxstream.NewDefaultNetworkKey()
	currentTimeProvider := mocks.NewCurrentTimeProviderMock()

	oneToTwoReader, oneToTwoWriter := io.Pipe()
	twoToOneReader, twoToOneWriter := io.Pipe()

	conn1 := newMockReadWriteCloser(twoToOneReader, oneToTwoWriter)
	defer conn1.Close()

	conn2 := newMockReadWriteCloser(oneToTwoReader, twoToOneWriter)
	defer conn2.Close()

	runTest(t, networkKey, currentTimeProvider, conn1, conn2)
}

func TestHandshaker_HandshakerSetsDeadlinesIfConnectionImplementsSetDeadliner(t *testing.T) {
	t.Parallel()

	networkKey := boxstream.NewDefaultNetworkKey()
	currentTimeProvider := mocks.NewCurrentTimeProviderMock()
	currentTimeProvider.CurrentTime = time.Now()

	oneToTwoReader, oneToTwoWriter := io.Pipe()
	twoToOneReader, twoToOneWriter := io.Pipe()

	conn1 := newReadWriteCloseSetDeadliner(twoToOneReader, oneToTwoWriter)
	defer conn1.Close()

	conn2 := newReadWriteCloseSetDeadliner(oneToTwoReader, twoToOneWriter)
	defer conn2.Close()

	runTest(t, networkKey, currentTimeProvider, conn1, conn2)

	require.Equal(t,
		[]time.Time{
			currentTimeProvider.CurrentTime.Add(15 * time.Second),
			{},
		},
		conn1.SetDeadlineCalls,
	)

	require.Empty(t, conn2.SetDeadlineCalls)
}

func runTest(t *testing.T, networkKey boxstream.NetworkKey, currentTimeProvider *mocks.CurrentTimeProviderMock, conn1 io.ReadWriteCloser, conn2 io.ReadWriteCloser) {
	peer1 := fixtures.SomePrivateIdentity()
	handshaker1, err := boxstream.NewHandshaker(peer1, networkKey, currentTimeProvider)
	require.NoError(t, err)

	peer2 := fixtures.SomePrivateIdentity()
	handshaker2, err := boxstream.NewHandshaker(peer2, networkKey, currentTimeProvider)
	require.NoError(t, err)

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

type mockReadWriteCloser struct {
	read  io.ReadCloser
	write io.WriteCloser
}

func newMockReadWriteCloser(read io.ReadCloser, write io.WriteCloser) *mockReadWriteCloser {
	return &mockReadWriteCloser{
		read:  read,
		write: write,
	}
}

func (m mockReadWriteCloser) Read(p []byte) (n int, err error) {
	return m.read.Read(p)
}

func (m mockReadWriteCloser) Write(p []byte) (n int, err error) {
	return m.write.Write(p)
}

func (m mockReadWriteCloser) Close() error {
	var err error
	err = multierror.Append(err, m.read.Close())
	err = multierror.Append(err, m.write.Close())
	return err
}

type mockReadWriteCloseSetDeadliner struct {
	*mockReadWriteCloser
	SetDeadlineCalls []time.Time
}

func newReadWriteCloseSetDeadliner(read io.ReadCloser, write io.WriteCloser) *mockReadWriteCloseSetDeadliner {
	return &mockReadWriteCloseSetDeadliner{
		mockReadWriteCloser: newMockReadWriteCloser(read, write),
	}
}

func (m *mockReadWriteCloseSetDeadliner) SetDeadline(t time.Time) error {
	m.SetDeadlineCalls = append(m.SetDeadlineCalls, t)
	return nil
}
