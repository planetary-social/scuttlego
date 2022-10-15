package rpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
	"github.com/stretchr/testify/require"
)

func TestResponseStreams_CancellingContextReleasesChannel(t *testing.T) {
	sender := NewSenderMock()
	logger := fixtures.SomeLogger()
	ctx := fixtures.TestContext(t)
	req := someRequest()

	streams := rpc.NewResponseStreams(sender, logger)

	stream, err := streams.Open(ctx, req)
	require.NoError(t, err)

	require.Len(t, sender.calls, 1, "opening the stream should send a request")

	go func() {
		response := createResponse(sender.calls[0])
		err := streams.HandleIncomingResponse(response)
		require.NoError(t, err)

		termination := createTermination(sender.calls[0])
		err = streams.HandleIncomingResponse(termination)
		require.NoError(t, err)
	}()

	select {
	case resp := <-stream.Channel():
		require.NoError(t, resp.Err)
	case <-time.After(5 * time.Second):
		t.Fatal("first response was not received")
	}

	select {
	case resp := <-stream.Channel():
		require.ErrorIs(t, resp.Err, rpc.ErrEndOrErr)
	case <-time.After(5 * time.Second):
		t.Fatal("channel was not released")
	}
}

func TestResponseStreams_RequestsAndInStreamMessagesAreCorrectlyMarshaledInDuplexStreams(t *testing.T) {
	sender := NewSenderMock()
	logger := fixtures.SomeLogger()

	ctx, cancel := context.WithCancel(fixtures.TestContext(t))

	req := rpc.MustNewRequest(
		fixtures.SomeProcedureName(),
		rpc.ProcedureTypeDuplex,
		fixtures.SomeJSON(),
	)
	inStreamMessageData := fixtures.SomeBytes()

	streams := rpc.NewResponseStreams(sender, logger)

	stream, err := streams.Open(ctx, req)
	require.NoError(t, err)

	err = stream.WriteMessage(inStreamMessageData)
	require.NoError(t, err)

	cancel()

	expectedRequestMessage := expectedFlagsAndRequestNumber{
		transport.MustNewMessageHeaderFlags(true, false, transport.MessageBodyTypeJSON),
		1,
	}
	expectedInStreamMessage := expectedFlagsAndRequestNumber{
		transport.MustNewMessageHeaderFlags(true, false, transport.MessageBodyTypeJSON),
		1,
	}
	expectedTerminationMessage := expectedFlagsAndRequestNumber{
		transport.MustNewMessageHeaderFlags(true, true, transport.MessageBodyTypeJSON),
		1,
	}
	expectedMessages := []expectedFlagsAndRequestNumber{
		expectedRequestMessage,
		expectedInStreamMessage,
		expectedTerminationMessage,
	}

	require.Eventually(t, func() bool {
		return len(sender.calls) == len(expectedMessages)
	}, time.Second, 10*time.Millisecond)

	for i := range expectedMessages {
		t.Log(i)

		expected := expectedMessages[i]
		actual := sender.calls[i]

		require.Equal(t, expected.Flags, actual.Header.Flags())
		require.Equal(t, expected.RequestNumber, actual.Header.RequestNumber())
	}
}

type expectedFlagsAndRequestNumber struct {
	Flags         transport.MessageHeaderFlags
	RequestNumber int
}

func TestResponseStreams_DoesNotAcceptProcedureOfTypeUnknown(t *testing.T) {
	testCases := []struct {
		Name          string
		ProcedureType rpc.ProcedureType
		ExpectedError error
	}{
		{
			Name:          "unknown",
			ProcedureType: rpc.ProcedureTypeUnknown,
			ExpectedError: errors.New("could not marshal a request: could not marshal the request body: could not encode the procedure type: unknown procedure type {s:unknown}"),
		},
		{
			Name:          "async",
			ProcedureType: rpc.ProcedureTypeAsync,
			ExpectedError: nil,
		},
		{
			Name:          "source",
			ProcedureType: rpc.ProcedureTypeSource,
			ExpectedError: nil,
		},
		{
			Name:          "duplex",
			ProcedureType: rpc.ProcedureTypeDuplex,
			ExpectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := fixtures.TestContext(t)
			sender := NewSenderMock()
			logger := fixtures.TestLogger(t)

			streams := rpc.NewResponseStreams(sender, logger)

			_, err := streams.Open(ctx,
				rpc.MustNewRequest(
					fixtures.SomeProcedureName(),
					testCase.ProcedureType,
					[]byte("{}"),
				))
			if testCase.ExpectedError != nil {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestResponseStreams_ReturnsImplementationForWhichWriteMessageWorksOnlyForDuplexStreams(t *testing.T) {
	testCases := []struct {
		Name          string
		ProcedureType rpc.ProcedureType
		ExpectedError bool
	}{
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
			logger := fixtures.TestLogger(t)

			streams := rpc.NewResponseStreams(sender, logger)

			stream, err := streams.Open(ctx,
				rpc.MustNewRequest(
					fixtures.SomeProcedureName(),
					testCase.ProcedureType,
					[]byte("{}"),
				))
			require.NoError(t, err)

			err = stream.WriteMessage(fixtures.SomeBytes())
			if testCase.ExpectedError {
				require.EqualError(t, err, "not a duplex stream")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func someRequest() *rpc.Request {
	req, err := rpc.NewRequest(
		fixtures.SomeProcedureName(),
		fixtures.SomeProcedureType(),
		fixtures.SomeJSON(),
	)
	if err != nil {
		panic(err)
	}

	return req
}

func createResponse(req *transport.Message) *transport.Message {
	content := fixtures.SomeBytes()

	header, err := transport.NewMessageHeader(
		transport.MustNewMessageHeaderFlags(fixtures.SomeBool(), false, transport.MessageBodyTypeJSON),
		uint32(len(content)),
		int32(-req.Header.RequestNumber()),
	)
	if err != nil {
		panic(err)
	}

	msg, err := transport.NewMessage(header, content)
	if err != nil {
		panic(err)
	}

	return &msg
}

func createTermination(req *transport.Message) *transport.Message {
	header, err := transport.NewMessageHeader(
		transport.MustNewMessageHeaderFlags(fixtures.SomeBool(), true, fixtures.SomeMessageBodyType()),
		0,
		int32(-req.Header.RequestNumber()),
	)
	if err != nil {
		panic(err)
	}

	msg, err := transport.NewMessage(header, nil)
	if err != nil {
		panic(err)
	}

	return &msg
}

type SenderMock struct {
	calls []*transport.Message
}

func NewSenderMock() *SenderMock {
	return &SenderMock{}
}

func (s *SenderMock) Send(msg *transport.Message) error {
	s.calls = append(s.calls, msg)
	return nil
}
