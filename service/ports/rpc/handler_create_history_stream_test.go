package rpc_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/internal/mocks"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	transportrpc "github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
	"github.com/planetary-social/scuttlego/service/ports/rpc"
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
			Keys:         internal.Ptr(true),
			ContainsKeys: true,
		},
		{
			Name:         "false",
			Keys:         internal.Ptr(false),
			ContainsKeys: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := fixtures.TestContext(t)
			queryHandler := NewMockCreateHistoryStreamQueryHandler()
			s := mocks.NewMockCloserStream()
			h := rpc.NewHandlerCreateHistoryStream(queryHandler, logging.NewDevNullLogger())

			req := createHistoryStreamRequest(t, testCase.Keys)

			h.Handle(ctx, s, req)
			require.Len(t, queryHandler.Calls, 1)

			msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())
			err := queryHandler.Calls[0].ResponseWriter.WriteMessage(msg)
			require.NoError(t, err)

			msgs := s.WrittenMessages()
			require.Len(t, msgs, 1)
			for _, msg := range msgs {
				keyJSON := []byte(`"key":`)
				require.Equal(t, testCase.ContainsKeys, bytes.Contains(msg.Body, keyJSON))
				require.Equal(t, transport.MessageBodyTypeJSON, msg.BodyType)
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

	req, err := messages.NewCreateHistoryStream(args)
	require.NoError(t, err)

	return req
}

type MockCreateHistoryStreamQueryHandler struct {
	Calls []queries.CreateHistoryStream
}

func NewMockCreateHistoryStreamQueryHandler() *MockCreateHistoryStreamQueryHandler {
	return &MockCreateHistoryStreamQueryHandler{}
}

func (m *MockCreateHistoryStreamQueryHandler) Handle(ctx context.Context, query queries.CreateHistoryStream) {
	m.Calls = append(m.Calls, query)
}
