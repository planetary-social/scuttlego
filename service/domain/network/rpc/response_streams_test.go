package rpc_test

import (
	"testing"
	"time"

	"github.com/planetary-social/go-ssb/fixtures"
	rpc2 "github.com/planetary-social/go-ssb/service/domain/network/rpc"
	"github.com/planetary-social/go-ssb/service/domain/network/rpc/transport"
	"github.com/stretchr/testify/require"
)

func TestCancellingContextReleasesChannel(t *testing.T) {
	sender := NewSenderMock()
	logger := fixtures.SomeLogger()
	ctx := fixtures.TestContext(t)
	req := someRequest()

	streams := rpc2.NewResponseStreams(sender, logger)

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
		require.ErrorIs(t, resp.Err, rpc2.ErrEndOrErr)
	case <-time.After(5 * time.Second):
		t.Fatal("channel was not released")
	}
}

func someRequest() *rpc2.Request {
	req, err := rpc2.NewRequest(
		fixtures.SomeProcedureName(),
		fixtures.SomeProcedureType(),
		fixtures.SomeBool(),
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
		transport.MessageHeaderFlags{},
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
		transport.MessageHeaderFlags{
			EndOrError: true,
		},
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
