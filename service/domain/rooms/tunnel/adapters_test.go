package tunnel_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/rooms/tunnel"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestResponseStreamReadWriteCloserAdapter_CloseCallsCancel(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()

	stream := newResponseStreamMock(ctx, nil)

	adapter := tunnel.NewResponseStreamReadWriteCloserAdapter(stream, cancel.Cancel)

	err := adapter.Close()
	require.NoError(t, err)
	require.True(t, cancel.Called)
}

func TestResponseStreamReadWriteCloserAdapter_ReadBlocksWaitingForData(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()
	stream := newResponseStreamMock(ctx, nil)
	adapter := tunnel.NewResponseStreamReadWriteCloserAdapter(stream, cancel.Cancel)

	ch := make(chan struct{})
	go func() {
		defer close(ch)
		adapter.Read(nil) //nolint:errcheck
	}()

	select {
	case <-ch:
		t.Fatal("read returned")
	case <-time.After(100 * time.Millisecond):
		t.Log("ok")
	}
}

func TestResponseStreamReadWriteCloserAdapter_ReadPropagatesStreamErrors(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()

	stream := newResponseStreamMock(ctx, []rpc.ResponseWithError{
		{
			Value: nil,
			Err:   errors.New("some error"),
		},
	})

	adapter := tunnel.NewResponseStreamReadWriteCloserAdapter(stream, cancel.Cancel)

	_, err := adapter.Read(nil)
	require.EqualError(t, err, "stream returned an error: some error")
}

func TestResponseStreamReadWriteCloserAdapter_SmallMessagesAreRetrievedInOneRead(t *testing.T) {
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

	adapter := tunnel.NewResponseStreamReadWriteCloserAdapter(stream, cancel.Cancel)

	buf := make([]byte, bufSize)

	n, err := adapter.Read(buf)
	require.NoError(t, err)
	require.Equal(t, len(msg), n)
	require.Equal(t, msg, buf[:n])
}

func TestResponseStreamReadWriteCloserAdapter_LargeMessagesAreRetrievedOverMultipleReads(t *testing.T) {
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

	adapter := tunnel.NewResponseStreamReadWriteCloserAdapter(stream, cancel.Cancel)

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

func TestResponseStreamReadWriteCloserAdapter_ReadWhenChannelIsClosedReturnsAnError(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()

	streamCtx, streamCtxCancel := context.WithCancel(ctx)
	streamCtxCancel()
	stream := newResponseStreamMock(streamCtx, nil)

	adapter := tunnel.NewResponseStreamReadWriteCloserAdapter(stream, cancel.Cancel)

	_, err := adapter.Read(nil)
	require.EqualError(t, err, "channel closed")
}

func TestResponseStreamReadWriteCloserAdapter_WriteCallsWriteMessage(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()
	stream := newResponseStreamMock(ctx, nil)
	adapter := tunnel.NewResponseStreamReadWriteCloserAdapter(stream, cancel.Cancel)

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

func TestStreamReadWriterCloserAdapter_CloseCallsCancel(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()
	stream := newStreamMock(ctx, nil)
	adapter := tunnel.NewStreamReadWriteCloserAdapter(stream, cancel.Cancel)

	err := adapter.Close()
	require.NoError(t, err)
	require.True(t, cancel.Called)
}

func TestStreamReadWriterCloserAdapter_ReadBlocksWaitingForData(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()
	stream := newStreamMock(ctx, nil)
	adapter := tunnel.NewStreamReadWriteCloserAdapter(stream, cancel.Cancel)

	ch := make(chan struct{})
	go func() {
		defer close(ch)
		adapter.Read(nil) //nolint:errcheck
	}()

	select {
	case <-ch:
		t.Fatal("read returned")
	case <-time.After(100 * time.Millisecond):
		t.Log("ok")
	}
}

func TestStreamReadWriterCloserAdapter_SmallMessagesAreRetrievedInOneRead(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()

	const bufSize = 100
	msg := bytes.Repeat([]byte("a"), bufSize/2)

	stream := newStreamMock(ctx, []rpc.IncomingMessage{
		{
			Body: msg,
		},
	})

	adapter := tunnel.NewStreamReadWriteCloserAdapter(stream, cancel.Cancel)

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

	stream := newStreamMock(ctx, []rpc.IncomingMessage{
		{
			Body: responseBytes,
		},
	})

	adapter := tunnel.NewStreamReadWriteCloserAdapter(stream, cancel.Cancel)

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
	stream := newStreamMock(streamCtx, nil)

	adapter := tunnel.NewStreamReadWriteCloserAdapter(stream, cancel.Cancel)

	_, err := adapter.Read(nil)
	require.EqualError(t, err, "channel closed")
}

func TestStreamReadWriterCloserAdapter_WriteCallsWriteMessage(t *testing.T) {
	ctx := fixtures.TestContext(t)
	cancel := newCancelFuncMock()
	stream := newStreamMock(ctx, nil)
	adapter := tunnel.NewStreamReadWriteCloserAdapter(stream, cancel.Cancel)

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

type cancelFuncMock struct {
	Called bool
}

func newCancelFuncMock() *cancelFuncMock {
	return &cancelFuncMock{}
}

func (m *cancelFuncMock) Cancel() {
	m.Called = true
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

type streamMock struct {
	ch                chan rpc.IncomingMessage
	WriteMessageCalls [][]byte
}

func newStreamMock(ctx context.Context, messagesToReceive []rpc.IncomingMessage) *streamMock {
	ch := make(chan rpc.IncomingMessage)
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
	return &streamMock{
		ch: ch,
	}
}

func (s *streamMock) IncomingMessages() (<-chan rpc.IncomingMessage, error) {
	return s.ch, nil
}

func (s *streamMock) WriteMessage(body []byte) error {
	s.WriteMessageCalls = append(s.WriteMessageCalls, body)
	return nil
}
