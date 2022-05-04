package boxstream_test

import (
	"io"
	"math/rand"
	"strings"
	"testing"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/domain/transport/boxstream"
	"github.com/stretchr/testify/require"
)

func TestStream_Small(t *testing.T) {
	data := []byte("test")
	stream1, stream2 := newStreams(t)
	testWriteRead(t, stream1, stream2, data)
	testWriteRead(t, stream2, stream1, data)
}

func TestStream_Big(t *testing.T) {
	data := []byte(strings.Repeat("test", 5000))
	stream1, stream2 := newStreams(t)
	testWriteRead(t, stream1, stream2, data)
	testWriteRead(t, stream2, stream1, data)
}

func testWriteRead(t *testing.T, stream1 *boxstream.Stream, stream2 *boxstream.Stream, data []byte) {
	go func() {
		n, err := stream1.Write(data)
		require.NoError(t, err)
		require.Equal(t, 0, n)
	}()

	buf := make([]byte, len(data))
	n, err := io.ReadFull(stream2, buf)
	require.NoError(t, err)
	require.Equal(t, len(data), n)
	require.Equal(t, data, buf)
}

func newStreams(t *testing.T) (*boxstream.Stream, *boxstream.Stream) {
	oneToTwoKey := someKey(t)
	oneToTwoNonce := someNonce(t)
	oneToTwoReader, oneToTwoWriter := io.Pipe()

	twoToOneKey := someKey(t)
	twoToOneNonce := someNonce(t)
	twoToOneReader, twoToOneWriter := io.Pipe()

	identity1 := fixtures.SomePublicIdentity()
	conn1 := newMockConnection(twoToOneReader, oneToTwoWriter)

	identity2 := fixtures.SomePublicIdentity()
	conn2 := newMockConnection(oneToTwoReader, twoToOneWriter)

	stream1, err := boxstream.NewStream(conn1, boxstream.HandshakeResult{
		Remote:      identity2,
		WriteSecret: oneToTwoKey,
		WriteNonce:  oneToTwoNonce,
		ReadSecret:  twoToOneKey,
		ReadNonce:   twoToOneNonce,
	})
	require.NoError(t, err)

	stream2, err := boxstream.NewStream(conn2, boxstream.HandshakeResult{
		Remote:      identity1,
		WriteSecret: twoToOneKey,
		WriteNonce:  twoToOneNonce,
		ReadSecret:  oneToTwoKey,
		ReadNonce:   oneToTwoNonce,
	})
	require.NoError(t, err)

	return stream1, stream2
}

func someKey(t *testing.T) boxstream.Key {
	var key boxstream.Key
	_, err := rand.Read(key[:])
	require.NoError(t, err)
	return key
}

func someNonce(t *testing.T) boxstream.Nonce {
	var nonce boxstream.Nonce
	_, err := rand.Read(nonce[:])
	require.NoError(t, err)
	return nonce
}
