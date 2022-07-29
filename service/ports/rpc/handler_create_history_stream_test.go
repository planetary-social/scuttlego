package rpc_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	transportrpc "github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux/mocks"
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
			rw := mocks.NewMockResponseWriterCloser()
			h := rpc.NewHandlerCreateHistoryStream(queryHandler, logging.NewDevNullLogger())

			req := createHistoryStreamRequest(t, testCase.Keys)

			h.Handle(ctx, rw, req)
			require.Len(t, queryHandler.Calls, 1)

			msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())
			err := queryHandler.Calls[0].ResponseWriter.WriteMessage(msg)
			require.NoError(t, err)
			require.Len(t, rw.WrittenMessages, 1)
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
