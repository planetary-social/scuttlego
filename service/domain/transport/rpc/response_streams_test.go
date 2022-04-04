package rpc_test

import (
	"testing"
	"time"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc/transport"
	"github.com/stretchr/testify/require"
)

func TestCancellingContextReleasesChannel(t *testing.T) {
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
