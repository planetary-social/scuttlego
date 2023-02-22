package rpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
	"github.com/stretchr/testify/require"
)

const (
	timeout = 1 * time.Second
)

func TestRequestStreams_RequestIsPassedToHandler(t *testing.T) {
	t.Parallel()

	data := []byte(`{"name":["some", "request"],"type":"async","args":["some", "args"]}`)

	msg, err := transport.NewMessage(
		transport.MustNewMessageHeader(
			transport.MustNewMessageHeaderFlags(
				fixtures.SomeBool(),
				false,
				transport.MessageBodyTypeJSON,
			),
			uint32(len(data)),
			123,
		),
		data,
	)
	require.NoError(t, err)

	sender := NewSenderMock()
	logger := fixtures.TestLogger(t)

	expectedRequest := rpc.MustNewRequest(
		rpc.MustNewProcedureName([]string{"some", "request"}),
		rpc.ProcedureTypeAsync,
		[]byte(`["some", "args"]`),
	)

	ok := make(chan struct{})
	handler := newRequestHandlerFunc(func(ctx context.Context, s rpc.Stream, req *rpc.Request) {
		require.Equal(t, expectedRequest, req)
		err := s.CloseWithError(nil)
		require.NoError(t, err)
		<-ctx.Done()
		close(ok)
	})

	ctx := fixtures.TestContext(t)
	streams := rpc.NewRequestStreams(sender, handler, logger)

	err = streams.HandleIncomingRequest(ctx, &msg)
	require.NoError(t, err)

	select {
	case <-time.After(timeout):
		t.Fatal("timeout, handler should have been called and then exited")
	case <-ok:
	}
}

func TestRequestStreams_ResponsesAreRejectedIfPassedByAccident(t *testing.T) {
	sender := NewSenderMock()
	logger := fixtures.TestLogger(t)
	ctx := fixtures.TestContext(t)
	streams := rpc.NewRequestStreams(sender, nil, logger)

	msg := someMessageWhichIsAResponse(t)

	err := streams.HandleIncomingRequest(ctx, &msg)
	require.EqualError(t, err, "passed a response")
}

func TestRequestStreams_PassingRequestsToClosedStreamsWillFailEarly(t *testing.T) {
	sender := NewSenderMock()
	logger := fixtures.TestLogger(t)
	ctx, cancel := context.WithCancel(fixtures.TestContext(t))
	streams := rpc.NewRequestStreams(sender, nil, logger)

	cancel()

	msg := someMessageOpeningAStream(t)

	err := streams.HandleIncomingRequest(ctx, &msg)
	require.EqualError(t, err, "context canceled")
}

func TestRequestStreams_PassingMalformedRequestIsNotAnError(t *testing.T) {
	t.Parallel()

	data := []byte(`malformed-data`)

	msg, err := transport.NewMessage(
		transport.MustNewMessageHeader(
			transport.MustNewMessageHeaderFlags(
				fixtures.SomeBool(),
				false,
				transport.MessageBodyTypeJSON,
			),
			uint32(len(data)),
			123,
		),
		data,
	)
	require.NoError(t, err)

	sender := NewSenderMock()
	logger := fixtures.TestLogger(t)

	ok := make(chan struct{})
	handler := newRequestHandlerFunc(func(ctx context.Context, s rpc.Stream, req *rpc.Request) {
		close(ok)
	})

	ctx := fixtures.TestContext(t)
	streams := rpc.NewRequestStreams(sender, handler, logger)

	err = streams.HandleIncomingRequest(ctx, &msg)
	require.NoError(t, err, "handling a malformed request should not be an error")

	select {
	case <-time.After(timeout):
		t.Log("timeout, handler correctly wasn't called (most likely)")
	case <-ok:
		t.Fatal("handler should not have been called")
	}
}

func TestRequestStreams_ClosingContextTerminatesCreatedRequestStreams(t *testing.T) {
	t.Parallel()

	sender := NewSenderMock()
	logger := fixtures.TestLogger(t)

	ok := make(chan struct{})
	handler := newRequestHandlerFunc(func(ctx context.Context, s rpc.Stream, req *rpc.Request) {
		<-ctx.Done()
		close(ok)
	})

	ctx, cancel := context.WithCancel(fixtures.TestContext(t))
	streams := rpc.NewRequestStreams(sender, handler, logger)
	msg := someMessageOpeningAStream(t)

	err := streams.HandleIncomingRequest(ctx, &msg)
	require.NoError(t, err)

	cancel()

	select {
	case <-time.After(timeout):
		t.Fatal("timeout, handler context should have been terminated but wasn't")
	case <-ok:
	}

	// a not very reliable way of documenting that if the handler doesn't close
	// the stream then the close message will not be sent
	<-time.After(timeout)
	require.Len(t, sender.SendCalls(), 0)
}

func TestRequestStreams_RemoteCanCloseRequestStreams(t *testing.T) {
	t.Parallel()

	sender := NewSenderMock()
	logger := fixtures.TestLogger(t)

	ok := make(chan struct{})
	handler := newRequestHandlerFunc(func(ctx context.Context, s rpc.Stream, req *rpc.Request) {
		<-ctx.Done()
		close(ok)
	})

	ctx := fixtures.TestContext(t)
	streams := rpc.NewRequestStreams(sender, handler, logger)

	msgOpen := someMessageOpeningAStream(t)
	msgTerminate := someMessageTerminatingAStream(t, msgOpen.Header.RequestNumber())

	err := streams.HandleIncomingRequest(ctx, &msgOpen)
	require.NoError(t, err)

	err = streams.HandleIncomingRequest(ctx, &msgTerminate)
	require.NoError(t, err)

	select {
	case <-time.After(timeout):
		t.Fatal("timeout, handler context should have been terminated but wasn't")
	case <-ok:
	}
}

func TestRequestStreams_HandlerCanCloseRequestStreams(t *testing.T) {
	t.Parallel()

	sender := NewSenderMock()
	logger := fixtures.TestLogger(t)

	ok := make(chan struct{})
	handler := newRequestHandlerFunc(func(ctx context.Context, s rpc.Stream, req *rpc.Request) {
		err := s.CloseWithError(nil)
		require.NoError(t, err)
		<-ctx.Done()
		close(ok)
	})

	ctx := fixtures.TestContext(t)
	streams := rpc.NewRequestStreams(sender, handler, logger)

	msgOpen := someMessageOpeningAStream(t)

	err := streams.HandleIncomingRequest(ctx, &msgOpen)
	require.NoError(t, err)

	select {
	case <-time.After(timeout):
		t.Fatal("handler should have exited after closing the stream")
	case <-ok:
	}
}

func TestRequestStreams_RemoteCanSendTerminationAfterTheStreamIsClosed(t *testing.T) {
	t.Parallel()

	sender := NewSenderMock()
	logger := fixtures.TestLogger(t)

	ok := make(chan struct{})
	handler := newRequestHandlerFunc(func(ctx context.Context, s rpc.Stream, req *rpc.Request) {
		err := s.CloseWithError(nil)
		require.NoError(t, err)
		<-ctx.Done()
		close(ok)
	})

	ctx := fixtures.TestContext(t)
	streams := rpc.NewRequestStreams(sender, handler, logger)

	msgOpen := someMessageOpeningAStream(t)
	err := streams.HandleIncomingRequest(ctx, &msgOpen)
	require.NoError(t, err)

	select {
	case <-time.After(timeout):
		t.Fatal("handler should have exited after closing the stream")
	case <-ok:
	}

	msgTermination := someMessageTerminatingAStream(t, msgOpen.Header.RequestNumber())
	err = streams.HandleIncomingRequest(ctx, &msgTermination)
	require.NoError(t, err)

	// It is hard to check that the message has been ignored so we are just
	// checking that receiving it doesn't return errors.
	// This test is overall not great as we don't know what happened:
	// - we may have hit the closed streams map
	// - we may have hit the termination function for a stream that still hasn't
	//   been cleaned up
}

func TestRequestStreams_MultipleMessagesSentInDuplexStream(t *testing.T) {
	t.Parallel()

	requestNumber := fixtures.SomePositiveInt32()

	requestMsgData := []byte(`{"name":["some", "request"],"type":"duplex","args":["some", "args"]}`)
	requestMsg, err := transport.NewMessage(
		transport.MustNewMessageHeader(
			transport.MustNewMessageHeaderFlags(
				fixtures.SomeBool(),
				false,
				transport.MessageBodyTypeJSON,
			),
			uint32(len(requestMsgData)),
			requestNumber,
		),
		requestMsgData,
	)
	require.NoError(t, err)

	streamMsgData := fixtures.SomeBytes()
	streamMsg, err := transport.NewMessage(
		transport.MustNewMessageHeader(
			transport.MustNewMessageHeaderFlags(
				fixtures.SomeBool(),
				false,
				transport.MessageBodyTypeJSON,
			),
			uint32(len(streamMsgData)),
			requestNumber,
		),
		streamMsgData,
	)
	require.NoError(t, err)

	closeStreamMsgData := fixtures.SomeBytes()
	closeStreamMsg, err := transport.NewMessage(
		transport.MustNewMessageHeader(
			transport.MustNewMessageHeaderFlags(
				fixtures.SomeBool(),
				true,
				transport.MessageBodyTypeJSON,
			),
			uint32(len(closeStreamMsgData)),
			requestNumber,
		),
		closeStreamMsgData,
	)
	require.NoError(t, err)

	sender := NewSenderMock()
	logger := fixtures.TestLogger(t)

	var request *rpc.Request
	var streamMessages []rpc.IncomingMessage

	ok := make(chan struct{})
	handler := newRequestHandlerFunc(func(ctx context.Context, s rpc.Stream, req *rpc.Request) {
		request = req

		ch, err := s.IncomingMessages()
		require.NoError(t, err)

		for msg := range ch {
			streamMessages = append(streamMessages, msg)
		}

		err = s.CloseWithError(nil)
		require.NoError(t, err)

		<-ctx.Done()
		close(ok)
	})

	ctx := fixtures.TestContext(t)
	streams := rpc.NewRequestStreams(sender, handler, logger)

	err = streams.HandleIncomingRequest(ctx, &requestMsg)
	require.NoError(t, err)

	err = streams.HandleIncomingRequest(ctx, &streamMsg)
	require.NoError(t, err)

	err = streams.HandleIncomingRequest(ctx, &closeStreamMsg)
	require.NoError(t, err)

	select {
	case <-time.After(timeout):
		t.Fatal("timeout, handler should have been called and then exited after being closed")
	case <-ok:
	}

	require.Equal(t,
		rpc.MustNewRequest(
			rpc.MustNewProcedureName([]string{"some", "request"}),
			rpc.ProcedureTypeDuplex,
			[]byte(`["some", "args"]`),
		),
		request)
	require.Equal(t,
		[]rpc.IncomingMessage{
			{
				Body: streamMsgData,
			},
		},
		streamMessages)
}

func someMessageOpeningAStream(t *testing.T) transport.Message {
	data := []byte(`{"name":["aaa"]}`)

	msg, err := transport.NewMessage(
		transport.MustNewMessageHeader(
			transport.MustNewMessageHeaderFlags(fixtures.SomeBool(), false, transport.MessageBodyTypeJSON),
			uint32(len(data)),
			fixtures.SomePositiveInt32(),
		),
		data,
	)
	require.NoError(t, err)

	return msg
}

func someMessageTerminatingAStream(t *testing.T, requestNumber int) transport.Message {
	data := fixtures.SomeBytes()

	msg, err := transport.NewMessage(
		transport.MustNewMessageHeader(
			transport.MustNewMessageHeaderFlags(fixtures.SomeBool(), true, transport.MessageBodyTypeJSON),
			uint32(len(data)),
			int32(requestNumber),
		),
		data,
	)
	require.NoError(t, err)

	return msg
}

func someMessageWhichIsAResponse(t *testing.T) transport.Message {
	data := fixtures.SomeBytes()

	msg, err := transport.NewMessage(
		transport.MustNewMessageHeader(
			transport.MustNewMessageHeaderFlags(fixtures.SomeBool(), false, transport.MessageBodyTypeJSON),
			uint32(len(data)),
			fixtures.SomeNegativeInt32(),
		),
		data,
	)
	require.NoError(t, err)

	return msg
}
