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

func TestRequestStream_WriteMessage(t *testing.T) {
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
		sender.calls,
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
				sender.calls,
			)
		})
	}
}

func TestRequestStream_CloseWithError_Twice(t *testing.T) {
	ctx := fixtures.TestContext(t)
	sender := NewSenderMock()

	requestNumber := 1

	stream, err := rpc.NewRequestStream(ctx, requestNumber, rpc.ProcedureTypeAsync, sender)
	require.NoError(t, err)

	err = stream.CloseWithError(nil)
	require.NoError(t, err)

	err = stream.CloseWithError(nil)
	require.EqualError(t, err, "already sent close stream")

	require.Len(t, sender.calls, 1)
}

func TestRequestStream_HandleNewMessage(t *testing.T) {
	testCases := []struct {
		Name                    string
		ProcedureType           rpc.ProcedureType
		AcceptsFollowUpMessages bool
	}{
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

			err = stream.HandleNewMessage(&msg)
			if testCase.AcceptsFollowUpMessages {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, "only duplex streams can receive more than one message")
			}
		})
	}
}
