package tunnel_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/rooms/tunnel"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDialer_DialViaRoomPerformsCorrectRequestsAndCallsInitializerWithCorrectArguments(t *testing.T) {
	clientPeerInitializer := newClientPeerInitializerMock()
	dialer := tunnel.NewDialer(clientPeerInitializer)

	ctx := fixtures.TestContext(t)
	portalConn := mocks.NewConnectionMock(ctx)
	portalPeerRef := fixtures.SomeRefIdentity()
	portalPeer := transport.MustNewPeer(
		portalPeerRef.Identity(),
		portalConn,
	)
	targetRef := fixtures.SomeRefIdentity()

	someResponseBytes := fixtures.SomeBytes()

	portalConn.Mock(func(req *rpc.Request) []rpc.ResponseWithError {
		assert.Equal(t, messages.TunnelConnectProcedure.Name(), req.Name())
		assert.Equal(t, messages.TunnelConnectProcedure.Typ(), req.Type())
		assert.Equal(t,
			fmt.Sprintf(`[{"portal":"%s","target":"%s"}]`, portalPeerRef.String(), targetRef.String()),
			string(req.Arguments()),
		)

		return []rpc.ResponseWithError{
			{
				Value: rpc.NewResponse(someResponseBytes),
				Err:   nil,
			},
		}
	})

	_, err := dialer.DialViaRoom(ctx, portalPeer, targetRef.Identity())
	require.NoError(t, err)

	require.Eventually(t,
		func() bool {
			if len(clientPeerInitializer.calls) > 0 {
				fmt.Println(targetRef.Identity().String(), clientPeerInitializer.calls[0].Remote.String())
				fmt.Println(len(someResponseBytes), len(clientPeerInitializer.calls[0].ReceivedMessage))
			}
			return assert.ObjectsAreEqual(clientPeerInitializer.calls, []clientPeerInitializerCall{
				{
					Remote:          targetRef.Identity(),
					ReceivedMessage: someResponseBytes,
				},
			})

		}, 1*time.Second, 10*time.Millisecond)
}

func TestStreamReadWriterCloserAdapter_CloseCallsCancel(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()

	stream := newResponseStreamMock(ctx, nil)

	adapter := tunnel.NewStreamReadWriterCloserAdapter(stream, cancel.Cancel)

	err := adapter.Close()
	require.NoError(t, err)
	require.True(t, cancel.Called)
}

func TestStreamReadWriterCloserAdapter_ReadBlocksWaitingForData(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()
	stream := newResponseStreamMock(ctx, nil)
	adapter := tunnel.NewStreamReadWriterCloserAdapter(stream, cancel.Cancel)

	ch := make(chan struct{})
	go func() {
		defer close(ch)
		_, err := adapter.Read(nil)
		require.EqualError(t, err, "stream returned an error: some error")
	}()

	select {
	case <-ch:
		t.Fatal("read returned")
	case <-time.After(100 * time.Millisecond):
		t.Log("ok")
	}
}

func TestStreamReadWriterCloserAdapter_ReadPropagatesStreamErrors(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()

	stream := newResponseStreamMock(ctx, []rpc.ResponseWithError{
		{
			Value: nil,
			Err:   errors.New("some error"),
		},
	})

	adapter := tunnel.NewStreamReadWriterCloserAdapter(stream, cancel.Cancel)

	_, err := adapter.Read(nil)
	require.EqualError(t, err, "stream returned an error: some error")
}

func TestStreamReadWriterCloserAdapter_SmallMessagesAreRetrievedInOneRead(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()

	const bufSize = 100
	msg := bytes.Repeat([]byte("a"), bufSize/2)

	stream := newResponseStreamMock(ctx, []rpc.ResponseWithError{
		{
			Value: rpc.NewResponse(msg),
			Err:   nil,
		},
	})

	adapter := tunnel.NewStreamReadWriterCloserAdapter(stream, cancel.Cancel)

	buf := make([]byte, bufSize)

	n, err := adapter.Read(buf)
	require.NoError(t, err)
	require.Equal(t, len(msg), n)
	require.Equal(t, msg, buf[:n])
}

func TestStreamReadWriterCloserAdapter_LargeMessagesAreRetrievedOverMultipleReads(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()

	const bufSize = 100

	first := bytes.Repeat([]byte("a"), bufSize)
	second := bytes.Repeat([]byte("b"), bufSize)

	var responseBytes []byte
	responseBytes = append(responseBytes, first...)
	responseBytes = append(responseBytes, second...)

	stream := newResponseStreamMock(ctx, []rpc.ResponseWithError{
		{
			Value: rpc.NewResponse(responseBytes),
			Err:   nil,
		},
	})

	adapter := tunnel.NewStreamReadWriterCloserAdapter(stream, cancel.Cancel)

	buf := make([]byte, bufSize)

	n, err := adapter.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 100, n)
	require.Equal(t, first, buf)

	n, err = adapter.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 100, n)
	require.Equal(t, second, buf)
}

func TestStreamReadWriterCloserAdapter_ReadWhenChannelIsClosedReturnsAnError(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()

	streamCtx, streamCtxCancel := context.WithCancel(ctx)
	streamCtxCancel()
	stream := newResponseStreamMock(streamCtx, nil)

	adapter := tunnel.NewStreamReadWriterCloserAdapter(stream, cancel.Cancel)

	_, err := adapter.Read(nil)
	require.EqualError(t, err, "channel closed")
}

func TestStreamReadWriterCloserAdapter_WriteCallsWriteMessage(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()
	stream := newResponseStreamMock(ctx, nil)
	adapter := tunnel.NewStreamReadWriterCloserAdapter(stream, cancel.Cancel)

	someBytes := fixtures.SomeBytes()

	n, err := adapter.Write(someBytes)
	require.NoError(t, err)
	require.Equal(t, len(someBytes), n)
	require.Equal(t,
		[][]byte{
			someBytes,
		},
		stream.WriteMessageCalls,
	)
}

type responseStreamMock struct {
	ch                chan rpc.ResponseWithError
	WriteMessageCalls [][]byte
}

func newResponseStreamMock(ctx context.Context, messagesToReceive []rpc.ResponseWithError) *responseStreamMock {
	ch := make(chan rpc.ResponseWithError)
	go func() {
		defer close(ch)

		for _, msgToReceive := range messagesToReceive {
			select {
			case ch <- msgToReceive:
				continue
			case <-ctx.Done():
				return
			}
		}

		<-ctx.Done()
	}()
	return &responseStreamMock{
		ch: ch,
	}
}

func (r *responseStreamMock) WriteMessage(body []byte) error {
	r.WriteMessageCalls = append(r.WriteMessageCalls, body)
	return nil
}

func (r *responseStreamMock) Channel() <-chan rpc.ResponseWithError {
	return r.ch
}

type cancelFuncMock struct {
	Called bool
}

func newCancelFuncMock() *cancelFuncMock {
	return &cancelFuncMock{}
}

func (m *cancelFuncMock) Cancel() {
	m.Called = true
}

type clientPeerInitializerMock struct {
	calls []clientPeerInitializerCall
}

func newClientPeerInitializerMock() *clientPeerInitializerMock {
	return &clientPeerInitializerMock{}
}

func (c *clientPeerInitializerMock) InitializeClientPeer(ctx context.Context, rwc io.ReadWriteCloser, remote identity.Public) (transport.Peer, error) {
	buf := make([]byte, 10000)
	n, err := rwc.Read(buf)
	if err != nil {
		return transport.Peer{}, err
	}

	c.calls = append(c.calls, clientPeerInitializerCall{
		ReceivedMessage: buf[:n],
		Remote:          remote,
	})

	return transport.Peer{}, nil
}

type clientPeerInitializerCall struct {
	Remote          identity.Public
	ReceivedMessage []byte
}
