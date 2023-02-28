package rpc_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
	"github.com/stretchr/testify/require"
)

func TestResponseStreams_RemoteTerminationReturnsErrRemoteEnd(t *testing.T) {
	sender := NewSenderMock()
	logger := fixtures.SomeLogger()
	ctx := fixtures.TestContext(t)
	req := someRequest()

	streams := rpc.NewResponseStreams(sender, logger)

	stream, err := streams.Open(ctx, req)
	require.NoError(t, err)

	require.Len(t, sender.SendCalls(), 1, "opening the stream should send a request")

	go func() {
		response := createResponse(sender.SendCalls()[0])
		err := streams.HandleIncomingResponse(response)
		require.NoError(t, err)

		termination := createCleanTermination(sender.SendCalls()[0])
		err = streams.HandleIncomingResponse(termination)
		require.NoError(t, err)
	}()

	select {
	case resp, ok := <-stream.Channel():
		require.True(t, ok)
		require.NoError(t, resp.Err)
	case <-time.After(5 * time.Second):
		t.Fatal("first response was not received")
	}

	select {
	case resp, ok := <-stream.Channel():
		require.True(t, ok)
		require.ErrorIs(t, resp.Err, rpc.ErrRemoteEnd)
	case <-time.After(5 * time.Second):
		t.Fatal("channel was not released")
	}
}

func TestResponseStreams_RemoteTerminationWithAnErrorReturnsErrRemoteError(t *testing.T) {
	sender := NewSenderMock()
	logger := fixtures.SomeLogger()
	ctx := fixtures.TestContext(t)
	req := someRequest()

	streams := rpc.NewResponseStreams(sender, logger)

	stream, err := streams.Open(ctx, req)
	require.NoError(t, err)

	require.Len(t, sender.SendCalls(), 1, "opening the stream should send a request")

	go func() {
		response := createResponse(sender.SendCalls()[0])
		err := streams.HandleIncomingResponse(response)
		require.NoError(t, err)

		termination := createErrorTermination(sender.SendCalls()[0])
		err = streams.HandleIncomingResponse(termination)
		require.NoError(t, err)
	}()

	select {
	case resp, ok := <-stream.Channel():
		require.True(t, ok)
		require.NoError(t, resp.Err)
	case <-time.After(5 * time.Second):
		t.Fatal("first response was not received")
	}

	select {
	case resp, ok := <-stream.Channel():
		require.True(t, ok)
		require.ErrorIs(t, resp.Err, &rpc.RemoteError{})
	case <-time.After(5 * time.Second):
		t.Fatal("channel was not released")
	}
}

func TestResponseStreams_ContextTerminationClosesTheChannel(t *testing.T) {
	sender := NewSenderMock()
	logger := fixtures.SomeLogger()
	ctx, cancel := context.WithCancel(fixtures.TestContext(t))
	req := someRequest()

	streams := rpc.NewResponseStreams(sender, logger)

	stream, err := streams.Open(ctx, req)
	require.NoError(t, err)

	require.Len(t, sender.SendCalls(), 1, "opening the stream should send a request")

	cancel()

	select {
	case _, ok := <-stream.Channel():
		require.False(t, ok)
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

	inStreamMessageBodyType := fixtures.SomeMessageBodyType()
	inStreamMessageData := fixtures.SomeBytes()

	streams := rpc.NewResponseStreams(sender, logger)

	stream, err := streams.Open(ctx, req)
	require.NoError(t, err)

	err = stream.WriteMessage(inStreamMessageData, inStreamMessageBodyType)
	require.NoError(t, err)

	cancel()

	expectedRequestMessage := expectedFlagsAndRequestNumber{
		transport.MustNewMessageHeaderFlags(true, false, transport.MessageBodyTypeJSON),
		1,
	}
	expectedInStreamMessage := expectedFlagsAndRequestNumber{
		transport.MustNewMessageHeaderFlags(true, false, inStreamMessageBodyType),
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
		return len(sender.SendCalls()) == len(expectedMessages)
	}, time.Second, 10*time.Millisecond)

	for i := range expectedMessages {
		t.Log(i)

		expected := expectedMessages[i]
		actual := sender.SendCalls()[i]

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

			err = stream.WriteMessage(fixtures.SomeBytes(), transport.MessageBodyTypeBinary)
			if testCase.ExpectedError {
				require.EqualError(t, err, "not a duplex stream")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestResponseStreams_StreamsShouldSendTerminationOnlyForCertainProcedureTypes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name                  string
		ProcedureType         rpc.ProcedureType
		ShouldSendTermination bool
	}{
		{
			Name:                  "source",
			ProcedureType:         rpc.ProcedureTypeSource,
			ShouldSendTermination: true,
		},
		{
			Name:                  "duplex",
			ProcedureType:         rpc.ProcedureTypeDuplex,
			ShouldSendTermination: true,
		},
		{
			Name:                  "async",
			ProcedureType:         rpc.ProcedureTypeAsync,
			ShouldSendTermination: false,
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			sender := NewSenderMock()
			logger := fixtures.SomeLogger()
			ctx, cancel := context.WithCancel(fixtures.TestContext(t))
			req := rpc.MustNewRequest(
				fixtures.SomeProcedureName(),
				testCase.ProcedureType,
				fixtures.SomeJSON(),
			)

			streams := rpc.NewResponseStreams(sender, logger)

			stream, err := streams.Open(ctx, req)
			require.NoError(t, err)

			require.Len(t, sender.SendCalls(), 1, "opening the stream should send a request")

			cancel()

			select {
			case _, ok := <-stream.Channel():
				require.False(t, ok)
			case <-time.After(5 * time.Second):
				t.Fatal("channel was not released")
			}

			if testCase.ShouldSendTermination {
				require.Eventually(t, func() bool {
					calls := sender.SendCalls()

					if len(calls) != 2 {
						return false
					}

					if !calls[1].Header.Flags().EndOrError() {
						return false
					}

					return true
				}, 1*time.Second, 10*time.Millisecond)
			} else {
				<-time.After(2 * time.Second)
				calls := sender.SendCalls()
				require.Equal(t, 1, len(calls))
				require.False(t, calls[0].Header.Flags().EndOrError())
			}
		})
	}
}

func TestErrRemoteError_Is(t *testing.T) {
	require.False(t, errors.Is(errors.New("some error"), rpc.RemoteError{}))
	require.True(t, errors.Is(rpc.NewRemoteError(nil), rpc.RemoteError{}))
	require.True(t, errors.Is(rpc.NewRemoteError(nil), &rpc.RemoteError{}))
	require.True(t, errors.Is(errors.Wrap(rpc.NewRemoteError(nil), "wrap"), rpc.RemoteError{}))
	require.True(t, errors.Is(errors.Wrap(rpc.NewRemoteError(nil), "wrap"), &rpc.RemoteError{}))
}

func TestErrRemoteError_As(t *testing.T) {
	v := fixtures.SomeBytes()

	var err1 rpc.RemoteError
	require.True(t, errors.As(rpc.NewRemoteError(v), &err1))
	require.Equal(t, v, err1.Response())

	var err2 rpc.RemoteError
	require.True(t, errors.As(errors.Wrap(rpc.NewRemoteError(v), "wrap"), &err2))
	require.Equal(t, v, err2.Response())

	var err3 someError
	require.False(t, errors.As(rpc.NewRemoteError(v), &err3))
}

type someError struct {
}

func (s someError) Error() string {
	panic("this is a test type and we should never call this function")
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

func createCleanTermination(req *transport.Message) *transport.Message {
	header, err := transport.NewMessageHeader(
		transport.MustNewMessageHeaderFlags(fixtures.SomeBool(), true, transport.MessageBodyTypeJSON),
		4,
		int32(-req.Header.RequestNumber()),
	)
	if err != nil {
		panic(err)
	}

	msg, err := transport.NewMessage(header, []byte("true"))
	if err != nil {
		panic(err)
	}

	return &msg
}

func createErrorTermination(req *transport.Message) *transport.Message {
	header, err := transport.NewMessageHeader(
		transport.MustNewMessageHeaderFlags(fixtures.SomeBool(), true, transport.MessageBodyTypeJSON),
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
	lock  sync.Mutex
}

func NewSenderMock() *SenderMock {
	return &SenderMock{}
}

func (s *SenderMock) Send(msg *transport.Message) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.calls = append(s.calls, msg)
	return nil
}

func (s *SenderMock) SendCalls() []*transport.Message {
	s.lock.Lock()
	defer s.lock.Unlock()

	tmp := make([]*transport.Message, len(s.calls))
	copy(tmp, s.calls)
	return tmp
}
