package rpc_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/app/queries"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/messages"
	transportrpc "github.com/planetary-social/go-ssb/service/domain/transport/rpc"
	"github.com/planetary-social/go-ssb/service/ports/rpc"
	"github.com/stretchr/testify/require"
)

func TestCreateHistoryStream(t *testing.T) {
	testCases := []struct {
		Name         string
		Keys         *bool
		ContainsKeys bool
	}{
		{
			Name:         "nil",
			Keys:         nil,
			ContainsKeys: true,
		},
		{
			Name:         "true",
			Keys:         boolPointer(true),
			ContainsKeys: true,
		},
		{
			Name:         "false",
			Keys:         boolPointer(false),
			ContainsKeys: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := fixtures.TestContext(t)
			queryHandler := NewMockCreateHistoryStreamQueryHandler()
			rw := NewMockResponseWriter()
			h := rpc.NewHandlerCreateHistoryStream(queryHandler)

			queryHandler.MessagesToSend = []message.Message{
				fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed()),
				fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed()),
			}

			req := createHistoryStreamRequest(t, testCase.Keys)

			err := h.Handle(ctx, rw, req)
			require.NoError(t, err)

			require.Len(t, rw.WrittenMessages, 2)
			for _, body := range rw.WrittenMessages {
				keyJSON := []byte(`"key":`)
				require.Equal(t, testCase.ContainsKeys, bytes.Contains(body, keyJSON))
			}
		})
	}

}

func createHistoryStreamRequest(t *testing.T, keys *bool) *transportrpc.Request {
	args, err := messages.NewCreateHistoryStreamArguments(
		fixtures.SomeRefFeed(),
		nil,
		nil,
		nil,
		nil,
		keys,
	)
	require.NoError(t, err)

	argsBytes, err := args.MarshalJSON()
	require.NoError(t, err)

	req, err := transportrpc.NewRequest(
		fixtures.SomeProcedureName(),
		fixtures.SomeProcedureType(),
		argsBytes,
	)
	require.NoError(t, err)

	return req
}

type MockResponseWriter struct {
	WrittenMessages [][]byte
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{}
}

func (m *MockResponseWriter) WriteMessage(body []byte) error {
	cpy := make([]byte, len(body))
	copy(cpy, body)
	m.WrittenMessages = append(m.WrittenMessages, cpy)
	return nil
}

type MockCreateHistoryStreamQueryHandler struct {
	MessagesToSend []message.Message
}

func NewMockCreateHistoryStreamQueryHandler() *MockCreateHistoryStreamQueryHandler {
	return &MockCreateHistoryStreamQueryHandler{}
}

func (m MockCreateHistoryStreamQueryHandler) Handle(ctx context.Context, query queries.CreateHistoryStream) <-chan queries.MessageWithErr {
	ch := make(chan queries.MessageWithErr)

	go func() {
		defer close(ch)

		for _, msg := range m.MessagesToSend {
			select {
			case ch <- queries.MessageWithErr{Message: msg}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

func boolPointer(v bool) *bool {
	return &v
}
