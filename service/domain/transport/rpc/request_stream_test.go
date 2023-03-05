package rpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
	"github.com/stretchr/testify/require"
)

func TestRequestStream_RequestNumberMustBePositive(t *testing.T) {
	ctx := fixtures.TestContext(t)
	onLocalClose := newOnLocalCloseMock()
	sender := NewSenderMock()

	_, err := rpc.NewRequestStream(ctx, onLocalClose.Fn, -123, rpc.ProcedureTypeAsync, sender)
	require.EqualError(t, err, "number must be positive")

	_, err = rpc.NewRequestStream(ctx, onLocalClose.Fn, 0, rpc.ProcedureTypeAsync, sender)
	require.EqualError(t, err, "number must be positive")
}

func TestRequestStream_RequestNumberIsReturnedFromTheGetter(t *testing.T) {
	ctx := fixtures.TestContext(t)
	onLocalClose := newOnLocalCloseMock()
	sender := NewSenderMock()

	requestNumber := fixtures.SomeNonNegativeInt()

	stream, err := rpc.NewRequestStream(ctx, onLocalClose.Fn, requestNumber, rpc.ProcedureTypeAsync, sender)
	require.NoError(t, err)

	require.Equal(t, requestNumber, stream.RequestNumber())
}

func TestRequestStream_WriteMessageCallsMessageSender(t *testing.T) {
	ctx := fixtures.TestContext(t)
	onLocalClose := newOnLocalCloseMock()
	sender := NewSenderMock()

	requestNumber := fixtures.SomeNonNegativeInt()

	stream, err := rpc.NewRequestStream(ctx, onLocalClose.Fn, requestNumber, rpc.ProcedureTypeAsync, sender)
	require.NoError(t, err)

	body := []byte("some body")
	bodyType := fixtures.SomeMessageBodyType()

	err = stream.WriteMessage(body, bodyType)
	require.NoError(t, err)

	expectedMessage, err := transport.NewMessage(
		transport.MustNewMessageHeader(
			transport.MustNewMessageHeaderFlags(true, false, bodyType),
			uint32(len(body)),
			int32(-requestNumber),
		),
		body,
	)
	require.NoError(t, err)

	require.Equal(t,
		[]*transport.Message{
			&expectedMessage,
		},
		sender.SendCalls(),
	)
}

func TestRequestStream_CloseWithErrorSendsMessageAndCallsOnLocalClose(t *testing.T) {
	testCases := []struct {
		Name         string
		Err          error
		ExpectedBody []byte
	}{
		{
			Name:         "nil_error",
			Err:          nil,
			ExpectedBody: []byte("true"),
		},
		{
			Name:         "not_nil_error",
			Err:          errors.New("some error"),
			ExpectedBody: []byte(`{"error":"some error"}`),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := fixtures.TestContext(t)
			onLocalClose := newOnLocalCloseMock()
			sender := NewSenderMock()

			requestNumber := 1

			stream, err := rpc.NewRequestStream(ctx, onLocalClose.Fn, requestNumber, rpc.ProcedureTypeAsync, sender)
			require.NoError(t, err)

			err = stream.CloseWithError(testCase.Err)
			require.NoError(t, err)

			expectedMessage, err := transport.NewMessage(
				transport.MustNewMessageHeader(
					transport.MustNewMessageHeaderFlags(true, true, transport.MessageBodyTypeJSON),
					uint32(len(testCase.ExpectedBody)),
					int32(-requestNumber),
				),
				testCase.ExpectedBody,
			)
			require.NoError(t, err)

			require.Equal(t,
				[]*transport.Message{
					&expectedMessage,
				},
				sender.SendCalls(),
			)

			require.Equal(t, 1, onLocalClose.Calls)
		})
	}
}

func TestRequestStream_CloseWithErrorReturnsAnErrorWhenCalledForTheSecondTime(t *testing.T) {
	ctx := fixtures.TestContext(t)
	onLocalClose := newOnLocalCloseMock()
	sender := NewSenderMock()

	requestNumber := 1

	stream, err := rpc.NewRequestStream(ctx, onLocalClose.Fn, requestNumber, rpc.ProcedureTypeAsync, sender)
	require.NoError(t, err)

	err = stream.CloseWithError(nil)
	require.NoError(t, err)

	err = stream.CloseWithError(nil)
	require.EqualError(t, err, "already sent close stream")

	require.Len(t, sender.SendCalls(), 1)
	require.Equal(t, 1, onLocalClose.Calls)
}

func TestRequestStream_HandleNewMessageReturnsAnErrorForProceduresThatAreNotDuplex(t *testing.T) {
	testCases := []struct {
		Name                    string
		ProcedureType           rpc.ProcedureType
		AcceptsFollowUpMessages bool
	}{
		{
			Name:                    "unknown",
			ProcedureType:           rpc.ProcedureTypeUnknown,
			AcceptsFollowUpMessages: false,
		},
		{
			Name:                    "async",
			ProcedureType:           rpc.ProcedureTypeAsync,
			AcceptsFollowUpMessages: false,
		},
		{
			Name:                    "source",
			ProcedureType:           rpc.ProcedureTypeSource,
			AcceptsFollowUpMessages: false,
		},
		{
			Name:                    "duplex",
			ProcedureType:           rpc.ProcedureTypeDuplex,
			AcceptsFollowUpMessages: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := fixtures.TestContext(t)
			onLocalClose := newOnLocalCloseMock()
			sender := NewSenderMock()

			requestNumber := fixtures.SomePositiveInt32()

			stream, err := rpc.NewRequestStream(ctx, onLocalClose.Fn, int(requestNumber), testCase.ProcedureType, sender)
			require.NoError(t, err)

			msg := someDuplexIncomingMessage(t, requestNumber)

			if testCase.AcceptsFollowUpMessages {
				ch, err := stream.IncomingMessages()
				require.NoError(t, err)

				go func() {
					for range ch {
					}
				}()
			}

			err = stream.HandleNewMessage(msg)
			if testCase.AcceptsFollowUpMessages {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, "only duplex streams can receive messages")
			}
		})
	}
}

func TestRequestStream_IncomingMessagesReturnsAnErrorForProceduresThatAreNotDuplex(t *testing.T) {
	testCases := []struct {
		Name          string
		ProcedureType rpc.ProcedureType
		ExpectedError bool
	}{
		{
			Name:          "unknown",
			ProcedureType: rpc.ProcedureTypeUnknown,
			ExpectedError: true,
		},
		{
			Name:          "async",
			ProcedureType: rpc.ProcedureTypeAsync,
			ExpectedError: true,
		},
		{
			Name:          "source",
			ProcedureType: rpc.ProcedureTypeSource,
			ExpectedError: true,
		},
		{
			Name:          "duplex",
			ProcedureType: rpc.ProcedureTypeDuplex,
			ExpectedError: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := fixtures.TestContext(t)
			onLocalClose := newOnLocalCloseMock()
			sender := NewSenderMock()

			stream, err := rpc.NewRequestStream(ctx, onLocalClose.Fn, fixtures.SomePositiveInt(), testCase.ProcedureType, sender)
			require.NoError(t, err)

			_, err = stream.IncomingMessages()
			if testCase.ExpectedError {
				require.EqualError(t, err, "only duplex streams can receive messages")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRequestStream_IncomingMessagesBlockUntilStreamIsClosed(t *testing.T) {
	ctx := fixtures.TestContext(t)
	onLocalClose := newOnLocalCloseMock()
	sender := NewSenderMock()

	requestNumber := fixtures.SomePositiveInt()

	stream, err := rpc.NewRequestStream(ctx, onLocalClose.Fn, requestNumber, rpc.ProcedureTypeDuplex, sender)
	require.NoError(t, err)

	ch, err := stream.IncomingMessages()
	require.NoError(t, err)

	select {
	case <-time.After(timeout):
		t.Log("ok, blocked for a while")
	case <-ch:
		t.Fatal("closed")
	}
}

func TestRequestStream_IncomingMessagesReturnsClosedChannelIfStreamIsClosed(t *testing.T) {
	messagesToReceive := []*transport.Message{
		{
			Header: transport.MustNewMessageHeader(
				transport.MustNewMessageHeaderFlags(true, false, transport.MessageBodyTypeJSON),
				fixtures.SomeUint32(),
				1,
			),
			Body: rpc.MustMarshalRequestBody(
				rpc.MustNewRequest(
					fixtures.SomeProcedureName(),
					rpc.ProcedureTypeDuplex,
					[]byte("[]"),
				),
			),
		},
	}

	handlerEndedProcessingCh := make(chan struct{})

	handler := newRequestHandlerFunc(func(ctx context.Context, s rpc.Stream, req *rpc.Request) {
		defer close(handlerEndedProcessingCh)

		err := s.CloseWithError(nil)
		require.NoError(t, err)

		require.Eventually(t,
			func() bool {
				ch, err := s.IncomingMessages()
				if err != nil {
					return false
				}

				select {
				case _, ok := <-ch:
					return ok == false
				default:
					return false
				}

			}, timeout, tick)
	})

	ts := newTestConnection(t, handler)
	go ts.Raw.ReceiveMessages(messagesToReceive...)

	select {
	case <-handlerEndedProcessingCh:
		t.Log("ok")
	case <-time.After(timeout):
		t.Fatal("timeout")
	}
}

func TestRequestStream_IncomingMessagesReceivesIncomingMessagesAndThenClosesWhenStreamCloses(t *testing.T) {
	requestNumber := fixtures.SomePositiveInt32()
	openingMessage := someDuplexOpeningMessage(t, requestNumber)
	incomingMessage := someDuplexIncomingMessage(t, requestNumber)

	messagesToReceive := []*transport.Message{
		&openingMessage,
		&incomingMessage,
	}

	handlerEndedProcessingCh := make(chan struct{})
	handlerGotOneIncomingMessageCh := make(chan struct{})

	var incomingMessages []rpc.IncomingMessage

	handler := newRequestHandlerFunc(func(ctx context.Context, s rpc.Stream, req *rpc.Request) {
		ch, err := s.IncomingMessages()
		require.NoError(t, err)

		go func() {
			defer close(handlerEndedProcessingCh)

			for v := range ch {
				if len(incomingMessages) == 0 {
					close(handlerGotOneIncomingMessageCh)
				}
				incomingMessages = append(incomingMessages, v)
			}
		}()

		<-handlerGotOneIncomingMessageCh
		err = s.CloseWithError(nil)
		require.NoError(t, err)
	})

	ts := newTestConnection(t, handler)
	go ts.Raw.ReceiveMessages(messagesToReceive...)

	select {
	case <-time.After(timeout):
		t.Fatal("timeout")
	case <-handlerEndedProcessingCh:
		t.Log("ok")
	}

	require.Equal(t,
		[]rpc.IncomingMessage{
			{
				Body: incomingMessage.Body,
			},
		},
		incomingMessages,
	)
}

func someDuplexOpeningMessage(t *testing.T, requestNumber int32) transport.Message {
	data := rpc.MustMarshalRequestBody(
		rpc.MustNewRequest(
			fixtures.SomeProcedureName(),
			rpc.ProcedureTypeDuplex,
			[]byte("[]"),
		),
	)

	msg, err := transport.NewMessage(
		transport.MustNewMessageHeader(
			transport.MustNewMessageHeaderFlags(fixtures.SomeBool(), false, transport.MessageBodyTypeJSON),
			uint32(len(data)),
			requestNumber,
		),
		data,
	)
	require.NoError(t, err)

	return msg
}

func someDuplexIncomingMessage(t *testing.T, requestNumber int32) transport.Message {
	data := fixtures.SomeBytes()

	msg, err := transport.NewMessage(
		transport.MustNewMessageHeader(
			transport.MustNewMessageHeaderFlags(fixtures.SomeBool(), false, transport.MessageBodyTypeJSON),
			uint32(len(data)),
			requestNumber,
		),
		data,
	)
	require.NoError(t, err)

	return msg
}

type onLocalCloseMock struct {
	Calls int
}

func newOnLocalCloseMock() *onLocalCloseMock {
	return &onLocalCloseMock{}
}

func (m *onLocalCloseMock) Fn(rs *rpc.RequestStream) {
	m.Calls++
}

type testConnection struct {
	Connection *rpc.Connection
	Ctx        context.Context
	Raw        *rawConnectionMock
}

func newTestConnection(tb testing.TB, handler rpc.RequestHandler) testConnection {
	ctx := fixtures.TestContext(tb)
	logger := fixtures.SomeLogger()
	raw := newRawConnectionMock()

	conn, err := rpc.NewConnection(fixtures.SomeConnectionId(), fixtures.SomeBool(), raw, handler, logger)
	require.NoError(tb, err)

	loopCtx, cancelLoopCtx := context.WithCancel(ctx)

	connectionLoopClosedCh := make(chan struct{})
	tb.Cleanup(func() {
		cancelLoopCtx()
		<-connectionLoopClosedCh
	})

	go func() {
		defer close(connectionLoopClosedCh)
		if err := conn.Loop(loopCtx); err != nil {
			logger.Debug().WithError(err).Message("conn loop exited")
		}
	}()

	tb.Cleanup(func() {
		conn.Close()
	})

	return testConnection{
		Connection: conn,
		Ctx:        ctx,
		Raw:        raw,
	}
}
