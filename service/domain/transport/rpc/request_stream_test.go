package rpc_test

import (
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
	"github.com/stretchr/testify/require"
)

func TestRequestStream_RequestNumberMustBePositive(t *testing.T) {
	ctx := fixtures.TestContext(t)
	sender := NewSenderMock()

	_, err := rpc.NewRequestStream(ctx, -123, rpc.ProcedureTypeAsync, sender)
	require.EqualError(t, err, "number must be positive")

	_, err = rpc.NewRequestStream(ctx, 0, rpc.ProcedureTypeAsync, sender)
	require.EqualError(t, err, "number must be positive")
}

func TestRequestStream_RequestNumber(t *testing.T) {
	ctx := fixtures.TestContext(t)
	sender := NewSenderMock()

	requestNumber := fixtures.SomeNonNegativeInt()

	stream, err := rpc.NewRequestStream(ctx, requestNumber, rpc.ProcedureTypeAsync, sender)
	require.NoError(t, err)

	require.Equal(t, requestNumber, stream.RequestNumber())
}

func TestRequestStream_WriteMessageCallsMessageSender(t *testing.T) {
	ctx := fixtures.TestContext(t)
	sender := NewSenderMock()

	requestNumber := fixtures.SomeNonNegativeInt()

	stream, err := rpc.NewRequestStream(ctx, requestNumber, rpc.ProcedureTypeAsync, sender)
	require.NoError(t, err)

	body := []byte("some body")

	err = stream.WriteMessage(body)
	require.NoError(t, err)

	expectedMessage, err := transport.NewMessage(
		transport.MustNewMessageHeader(
			transport.MustNewMessageHeaderFlags(true, false, transport.MessageBodyTypeJSON),
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

func TestRequestStream_TerminatedByRemoteCanBeCalledManyTimes(t *testing.T) {
	ctx := fixtures.TestContext(t)
	sender := NewSenderMock()

	stream, err := rpc.NewRequestStream(ctx, fixtures.SomeNonNegativeInt(), rpc.ProcedureTypeAsync, sender)
	require.NoError(t, err)

	stream.TerminatedByRemote()
	select {
	case <-stream.Context().Done():
		t.Log("ok, stream is closed ")
	case <-time.After(timeout):
		t.Fatal("timeout, stream is not closed")
	}

	stream.TerminatedByRemote()
	select {
	case <-stream.Context().Done():
		t.Log("ok, stream is closed ")
	case <-time.After(timeout):
		t.Fatal("timeout, stream is not closed")
	}
}

func TestRequestStream_CloseWithError(t *testing.T) {
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
			sender := NewSenderMock()

			requestNumber := 1

			stream, err := rpc.NewRequestStream(ctx, requestNumber, rpc.ProcedureTypeAsync, sender)
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
		})
	}
}

func TestRequestStream_CloseWithErrorReturnsAnErrorWhenCalledForTheSecondTime(t *testing.T) {
	ctx := fixtures.TestContext(t)
	sender := NewSenderMock()

	requestNumber := 1

	stream, err := rpc.NewRequestStream(ctx, requestNumber, rpc.ProcedureTypeAsync, sender)
	require.NoError(t, err)

	err = stream.CloseWithError(nil)
	require.NoError(t, err)

	err = stream.CloseWithError(nil)
	require.EqualError(t, err, "already sent close stream")

	require.Len(t, sender.SendCalls(), 1)
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
			sender := NewSenderMock()

			requestNumber := 1

			stream, err := rpc.NewRequestStream(ctx, requestNumber, testCase.ProcedureType, sender)
			require.NoError(t, err)

			msg := someDuplexIncomingMessage(t, requestNumber)

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
			sender := NewSenderMock()

			stream, err := rpc.NewRequestStream(ctx, fixtures.SomePositiveInt(), testCase.ProcedureType, sender)
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
	sender := NewSenderMock()

	requestNumber := fixtures.SomePositiveInt()

	stream, err := rpc.NewRequestStream(ctx, requestNumber, rpc.ProcedureTypeDuplex, sender)
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
	ctx := fixtures.TestContext(t)
	sender := NewSenderMock()

	requestNumber := fixtures.SomePositiveInt()

	stream, err := rpc.NewRequestStream(ctx, requestNumber, rpc.ProcedureTypeDuplex, sender)
	require.NoError(t, err)

	err = stream.CloseWithError(nil)
	require.NoError(t, err)

	require.Eventually(t,
		func() bool {
			ch, err := stream.IncomingMessages()
			if err != nil {
				return false
			}

			select {
			case _, ok := <-ch:
				return ok == false
			default:
				return false
			}

		}, timeout, 10*time.Millisecond)
}

func TestRequestStream_IncomingMessagesReceivesIncomingMessages(t *testing.T) {
	ctx := fixtures.TestContext(t)
	sender := NewSenderMock()

	requestNumber := fixtures.SomePositiveInt()

	stream, err := rpc.NewRequestStream(ctx, requestNumber, rpc.ProcedureTypeDuplex, sender)
	require.NoError(t, err)

	ch, err := stream.IncomingMessages()
	require.NoError(t, err)

	var incomingMessages []rpc.IncomingMessage

	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		for v := range ch {
			incomingMessages = append(incomingMessages, v)
		}
	}()

	msg := someDuplexIncomingMessage(t, requestNumber)

	err = stream.HandleNewMessage(msg)
	require.NoError(t, err)

	err = stream.CloseWithError(nil)
	require.NoError(t, err)

	select {
	case <-time.After(timeout):
		t.Fatal("timeout")
	case <-doneCh:
		t.Log("ok, the channel was closed")
	}

	require.Equal(t,
		[]rpc.IncomingMessage{
			{
				Body: msg.Body,
			},
		},
		incomingMessages,
	)
}

func someDuplexIncomingMessage(t *testing.T, requestNumber int) transport.Message {
	data := fixtures.SomeBytes()

	msg, err := transport.NewMessage(
		transport.MustNewMessageHeader(
			transport.MustNewMessageHeaderFlags(fixtures.SomeBool(), false, transport.MessageBodyTypeJSON),
			uint32(len(data)),
			int32(requestNumber),
		),
		data,
	)
	require.NoError(t, err)

	return msg
}
