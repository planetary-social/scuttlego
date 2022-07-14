package queries_test

import (
	"context"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/stretchr/testify/require"
)

const createHistoryStreamTestDelay = 1 * time.Second

func TestCreateHistoryStream_IfOldAndLiveAreNotSetNothingIsWrittenAndStreamIsClosed(t *testing.T) {
	t.Parallel()

	ctx := fixtures.TestContext(t)
	a := makeQueriesAndRunCreateHistoryStreamHandler(t, ctx)

	rw := newCreateHistoryStreamResponseWriterMock()

	query := queries.CreateHistoryStream{
		Id:             fixtures.SomeRefFeed(),
		Seq:            nil,
		Limit:          nil,
		Live:           false,
		Old:            false,
		ResponseWriter: rw,
	}

	a.Queries.CreateHistoryStream.Handle(ctx, query)

	<-time.After(createHistoryStreamTestDelay)

	require.Equal(t, 1, a.MessagePubSub.CallsCount, "live should have been subscribed to in the background")
	require.Empty(t, a.FeedRepository.GetMessagesCalls, "since old is not specified repository shouldn't have been called")
	require.Empty(t, rw.WrittenMessages)
	require.Equal(t, []error{nil}, rw.WrittenErrors)
}

func TestCreateHistoryStream_IfRespositoryReturnsAnErrorStreamIsClosed(t *testing.T) {
	t.Parallel()

	ctx := fixtures.TestContext(t)
	a := makeQueriesAndRunCreateHistoryStreamHandler(t, ctx)

	feed := fixtures.SomeRefFeed()
	rw := newCreateHistoryStreamResponseWriterMock()

	a.FeedRepository.GetMessagesReturnErr = errors.New("forced error")

	query := queries.CreateHistoryStream{
		Id:             feed,
		Seq:            nil,
		Limit:          nil,
		Live:           true,
		Old:            true,
		ResponseWriter: rw,
	}

	a.Queries.CreateHistoryStream.Handle(ctx, query)

	// it is basically impossible to correctly check if live messages will be
	// returned as it is impossible to check if something never happens

	<-time.After(createHistoryStreamTestDelay)

	require.Equal(t, 1, a.MessagePubSub.CallsCount, "live should have been subscribed to in the background")
	require.NotEmpty(t, a.FeedRepository.GetMessagesCalls)
	require.Empty(t, rw.WrittenMessages)
	require.Len(t, rw.WrittenErrors, 1)
	require.EqualError(t, rw.WrittenErrors[0], "could not retrieve messages: forced error")
}

//
//func TestCreateHistoryStream_repository_is_called_correctly(t *testing.T) {
//	feed := fixtures.SomeRefFeed()
//	seq := fixtures.SomeSequence()
//	limit := 10
//
//	testCases := []struct {
//		Name          string
//		Seq           *message.Sequence
//		Limit         *int
//		ExpectedCalls []mocks.FeedRepositoryMockGetMessagesCall
//	}{
//		{
//			Name:  "nil",
//			Seq:   nil,
//			Limit: nil,
//			ExpectedCalls: []mocks.FeedRepositoryMockGetMessagesCall{
//				{
//					Id:    feed,
//					Seq:   nil,
//					Limit: nil,
//				},
//			},
//		},
//		{
//			Name:  "values",
//			Seq:   &seq,
//			Limit: &limit,
//			ExpectedCalls: []mocks.FeedRepositoryMockGetMessagesCall{
//				{
//					Id:    feed,
//					Seq:   &seq,
//					Limit: &limit,
//				},
//			},
//		},
//	}
//
//	for _, testCase := range testCases {
//		t.Run(testCase.Name, func(t *testing.T) {
//			a, err := di.BuildTestQueries()
//			require.NoError(t, err)
//
//			ctx := fixtures.TestContext(t)
//
//			expectedMessages := []message.Message{
//				fixtures.SomeMessage(message.MustNewSequence(1), feed),
//				fixtures.SomeMessage(message.MustNewSequence(2), feed),
//				fixtures.SomeMessage(message.MustNewSequence(3), feed),
//			}
//
//			a.FeedRepository.GetMessagesReturnValue = expectedMessages
//
//			query := queries.CreateHistoryStream{
//				Id:    feed,
//				Seq:   testCase.Seq,
//				Limit: testCase.Limit,
//				Live:  false,
//				Old:   true,
//			}
//
//			ch := a.Queries.CreateHistoryStream.Handle(ctx, query)
//
//			var receivedMessages []message.Message
//			for msgWithError := range ch {
//				require.NoError(t, msgWithError.Err)
//				receivedMessages = append(receivedMessages, msgWithError.Message)
//			}
//
//			require.Equal(t, testCase.ExpectedCalls, a.FeedRepository.GetMessagesCalls)
//			require.Equal(t, expectedMessages, receivedMessages)
//		})
//	}
//}
//
//func TestCreateHistoryStream(t *testing.T) {
//	a, err := di.BuildTestQueries()
//	require.NoError(t, err)
//
//	ctx := fixtures.TestContext(t)
//
//	feed := fixtures.SomeRefFeed()
//
//	msg1 := fixtures.SomeMessage(message.MustNewSequence(1), feed)
//	msg2 := fixtures.SomeMessage(message.MustNewSequence(2), feed)
//	msg3 := fixtures.SomeMessage(message.MustNewSequence(3), feed)
//
//	testCases := []struct {
//		Name string
//
//		Seq   *message.Sequence
//		Limit *int
//
//		Repository []message.Message
//		PubSub     []message.Message
//
//		ExpectedMessages []message.Message
//	}{
//		{
//			Name:  "only_repository",
//			Seq:   nil,
//			Limit: nil,
//			Repository: []message.Message{
//				msg1,
//				msg2,
//			},
//			PubSub: nil,
//			ExpectedMessages: []message.Message{
//				msg1,
//				msg2,
//			},
//		},
//		{
//			Name:       "only_pubsub",
//			Seq:        nil,
//			Limit:      nil,
//			Repository: nil,
//			PubSub: []message.Message{
//				msg1,
//				msg2,
//			},
//			ExpectedMessages: []message.Message{
//				msg1,
//				msg2,
//			},
//		},
//		{
//			Name:  "pubsub_and_repository",
//			Seq:   nil,
//			Limit: nil,
//			Repository: []message.Message{
//				msg1,
//			},
//			PubSub: []message.Message{
//				msg2,
//			},
//			ExpectedMessages: []message.Message{
//				msg1,
//				msg2,
//			},
//		},
//		{
//			Name:  "pubsub_and_repository_with_overlapping_messages",
//			Seq:   nil,
//			Limit: nil,
//			Repository: []message.Message{
//				msg1,
//				msg2,
//			},
//			PubSub: []message.Message{
//				msg2,
//				msg3,
//			},
//			ExpectedMessages: []message.Message{
//				msg1,
//				msg2,
//				msg3,
//			},
//		},
//		{
//			Name:  "repository_should_enforce_sequence_by_itself",
//			Seq:   internal.Ptr(message.MustNewSequence(2)),
//			Limit: nil,
//			Repository: []message.Message{
//				msg1,
//				msg2,
//				msg3,
//			},
//			PubSub: nil,
//			ExpectedMessages: []message.Message{
//				msg1,
//				msg2,
//				msg3,
//			},
//		},
//		{
//			Name:  "earlier_messages_from_pubsub_should_be_omitted_if_repository_returned_something",
//			Seq:   internal.Ptr(message.MustNewSequence(1)),
//			Limit: nil,
//			Repository: []message.Message{
//				msg2,
//			},
//			PubSub: []message.Message{
//				msg1,
//				msg2,
//				msg3,
//			},
//			ExpectedMessages: []message.Message{
//				msg2,
//				msg3,
//			},
//		},
//		{
//			Name:       "earlier_messages_from_pubsub_should_be_omitted_if_repository_returned_nothing",
//			Seq:        internal.Ptr(message.MustNewSequence(1)),
//			Limit:      nil,
//			Repository: nil,
//			PubSub: []message.Message{
//				msg1,
//				msg2,
//				msg3,
//			},
//			ExpectedMessages: []message.Message{
//				msg2,
//				msg3,
//			},
//		},
//		{
//			Name:  "repository_should_enforce_limit_by_itself",
//			Limit: internal.Ptr(2),
//			Repository: []message.Message{
//				msg1,
//				msg2,
//				msg3,
//			},
//			PubSub: nil,
//			ExpectedMessages: []message.Message{
//				msg1,
//				msg2,
//				msg3,
//			},
//		},
//		{
//			Name:  "if_repository_exhausted_the_limit_then_live_should_do_nothing",
//			Limit: internal.Ptr(2),
//			Repository: []message.Message{
//				msg1,
//				msg2,
//			},
//			PubSub: []message.Message{
//				msg3,
//			},
//			ExpectedMessages: []message.Message{
//				msg1,
//				msg2,
//			},
//		},
//		{
//			Name:  "repository_should_count_towards_the_limit",
//			Limit: internal.Ptr(2),
//			Repository: []message.Message{
//				msg1,
//			},
//			PubSub: []message.Message{
//				msg2,
//				msg3,
//			},
//			ExpectedMessages: []message.Message{
//				msg1,
//				msg2,
//			},
//		},
//		{
//			Name:       "live_should_keep_track_of_the_limit",
//			Limit:      internal.Ptr(2),
//			Repository: nil,
//			PubSub: []message.Message{
//				msg1,
//				msg2,
//				msg3,
//			},
//			ExpectedMessages: []message.Message{
//				msg1,
//				msg2,
//			},
//		},
//		{
//			Name:  "if_repository_returns_messages_out_of_order_we_do_not_check_it",
//			Seq:   nil,
//			Limit: nil,
//			Repository: []message.Message{
//				msg3,
//				msg2,
//				msg1,
//			},
//			PubSub: nil,
//			ExpectedMessages: []message.Message{
//				msg3,
//				msg2,
//				msg1,
//			},
//		},
//		{
//			Name:       "pubsub_is_not_required_to_handle_messages_out_of_order",
//			Seq:        nil,
//			Limit:      nil,
//			Repository: nil,
//			PubSub: []message.Message{
//				msg3,
//				msg2,
//				msg1,
//			},
//			ExpectedMessages: []message.Message{
//				msg3,
//			},
//		},
//	}
//
//	for _, testCase := range testCases {
//		t.Run(testCase.Name, func(t *testing.T) {
//			a.FeedRepository.GetMessagesReturnValue = testCase.Repository
//			a.MessagePubSub.NewMessagesToSend = testCase.PubSub
//
//			query := queries.CreateHistoryStream{
//				Id:    feed,
//				Seq:   testCase.Seq,
//				Limit: testCase.Limit,
//				Live:  true,
//				Old:   true,
//			}
//
//			ch := a.Queries.CreateHistoryStream.Handle(ctx, query)
//
//			var receivedMessages []message.Message
//			for msgWithError := range ch {
//				if msgWithError.Err != nil {
//					continue
//				}
//
//				receivedMessages = append(receivedMessages, msgWithError.Message)
//			}
//
//			require.Equal(t, testCase.ExpectedMessages, receivedMessages)
//		})
//	}
//}

//func TestCreateHistoryStream(t *testing.T) {
//	testCases := []struct {
//		Name         string
//		Keys         *bool
//		ContainsKeys bool
//	}{
//		{
//			Name:         "nil",
//			Keys:         nil,
//			ContainsKeys: true,
//		},
//		{
//			Name:         "true",
//			Keys:         internal.Ptr(true),
//			ContainsKeys: true,
//		},
//		{
//			Name:         "false",
//			Keys:         internal.Ptr(false),
//			ContainsKeys: false,
//		},
//	}
//
//	for _, testCase := range testCases {
//		t.Run(testCase.Name, func(t *testing.T) {
//			ctx := fixtures.TestContext(t)
//			queryHandler := NewMockCreateHistoryStreamQueryHandler()
//			rw := NewMockResponseWriter()
//			h := rpc.NewHandlerCreateHistoryStream(queryHandler)
//
//			queryHandler.MessagesToSend = []message.Message{
//				fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed()),
//				fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed()),
//			}
//
//			req := createHistoryStreamRequest(t, testCase.Keys)
//
//			err := h.Handle(ctx, rw, req)
//			require.NoError(t, err)
//
//			require.Len(t, rw.WrittenMessages, 2)
//			for _, body := range rw.WrittenMessages {
//				keyJSON := []byte(`"key":`)
//				require.Equal(t, testCase.ContainsKeys, bytes.Contains(body, keyJSON))
//			}
//		})
//	}
//
//}
//
//func createHistoryStreamRequest(t *testing.T, keys *bool) *transportrpc.Request {
//	args, err := messages.NewCreateHistoryStreamArguments(
//		fixtures.SomeRefFeed(),
//		nil,
//		nil,
//		nil,
//		nil,
//		keys,
//	)
//	require.NoError(t, err)
//
//	req, err := messages.NewCreateHistoryStream(args)
//	require.NoError(t, err)
//
//	return req
//}
//
//type MockResponseWriter struct {
//	WrittenMessages [][]byte
//}
//
//func NewMockResponseWriter() *MockResponseWriter {
//	return &MockResponseWriter{}
//}
//
//func (m *MockResponseWriter) WriteMessage(body []byte) error {
//	cpy := make([]byte, len(body))
//	copy(cpy, body)
//	m.WrittenMessages = append(m.WrittenMessages, cpy)
//	return nil
//}
//
//type MockCreateHistoryStreamQueryHandler struct {
//	MessagesToSend []message.Message
//}
//
//func NewMockCreateHistoryStreamQueryHandler() *MockCreateHistoryStreamQueryHandler {
//	return &MockCreateHistoryStreamQueryHandler{}
//}
//
//func (m MockCreateHistoryStreamQueryHandler) Handle(ctx context.Context, query queries.CreateHistoryStream) <-chan queries.MessageWithErr {
//	ch := make(chan queries.MessageWithErr)
//
//	go func() {
//		defer close(ch)
//
//		for _, msg := range m.MessagesToSend {
//			select {
//			case ch <- queries.MessageWithErr{Message: msg}:
//			case <-ctx.Done():
//				return
//			}
//		}
//	}()
//
//	return ch
//}

func TestQueue_GetFromEmptyQueueShouldReturnNoValues(t *testing.T) {
	q := queries.NewRequestQueue()
	_, ok := q.Get()
	require.False(t, ok)
}

func TestQueue_GetShouldReturnValuesAddedWithAddAndClearTheQueue(t *testing.T) {
	v := queries.NewCreateHistoryStreamToProcess(
		fixtures.TestContext(t),
		queries.CreateHistoryStream{
			Id: fixtures.SomeRefFeed(),
		},
	)

	q := queries.NewRequestQueue()

	q.Add(v)

	retrievedValue, ok := q.Get()
	require.True(t, ok)
	require.Equal(t, v, retrievedValue)

	_, ok = q.Get()
	require.False(t, ok)
}

func makeQueriesAndRunCreateHistoryStreamHandler(t *testing.T, ctx context.Context) di.TestQueries {
	a, err := di.BuildTestQueries(t)
	require.NoError(t, err)
	go func() {
		if err := a.Queries.CreateHistoryStream.Run(ctx); err != nil {
			t.Log("run error", err)
		}
	}()
	return a
}

type createHistoryStreamResponseWriterMock struct {
	WrittenMessages []message.Message
	WrittenErrors   []error
}

func newCreateHistoryStreamResponseWriterMock() *createHistoryStreamResponseWriterMock {
	return &createHistoryStreamResponseWriterMock{}
}

func (c *createHistoryStreamResponseWriterMock) WriteMessage(msg message.Message) error {
	c.WrittenMessages = append(c.WrittenMessages, msg)
	return nil
}

func (c *createHistoryStreamResponseWriterMock) CloseWithError(err error) error {
	c.WrittenErrors = append(c.WrittenErrors, err)
	return nil
}
