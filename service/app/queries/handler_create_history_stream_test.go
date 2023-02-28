package queries_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const createHistoryStreamTestDelay = 1 * time.Second

func TestCreateHistoryStream_SubscribesToNewMessagesToProcessThem(t *testing.T) {
	t.Parallel()

	ctx := fixtures.TestContext(t)
	a := makeQueriesAndRunCreateHistoryStreamHandler(t, ctx)

	require.Eventually(t, func() bool {
		return a.MessagePubSub.SubscribeToNewMessagesCallsCount() == 1
	}, createHistoryStreamTestDelay, 10*time.Millisecond)
}

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

	require.Eventually(t, func() bool {
		return assert.ObjectsAreEqual([]error{nil}, rw.WrittenErrors())
	}, createHistoryStreamTestDelay, 10*time.Millisecond)

	require.Empty(t, a.FeedRepository.GetMessagesCalls(), "since old is not specified repository shouldn't have been called")
	require.Empty(t, rw.WrittenMessages())
}

func TestCreateHistoryStream_IfOldIsSetAndThereAreNoOldMessagesNothingIsWrittenAndStreamIsClosed(t *testing.T) {
	t.Parallel()

	ctx := fixtures.TestContext(t)
	a := makeQueriesAndRunCreateHistoryStreamHandler(t, ctx)

	rw := newCreateHistoryStreamResponseWriterMock()

	query := queries.CreateHistoryStream{
		Id:             fixtures.SomeRefFeed(),
		Seq:            nil,
		Limit:          nil,
		Live:           false,
		Old:            true,
		ResponseWriter: rw,
	}

	a.Queries.CreateHistoryStream.Handle(ctx, query)

	require.Eventually(t, func() bool {
		return assert.ObjectsAreEqual([]error{nil}, rw.WrittenErrors())
	}, createHistoryStreamTestDelay, 10*time.Millisecond)

	require.NotEmpty(t, a.FeedRepository.GetMessagesCalls())
	require.Empty(t, rw.WrittenMessages())
}

func TestCreateHistoryStream_IfRepositoryReturnsAnErrorStreamIsClosed(t *testing.T) {
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

	require.Eventually(t, func() bool {
		if l := len(rw.WrittenErrors()); l != 1 {
			t.Logf("length of written errors: %d", l)
			return false
		}

		if err := rw.WrittenErrors()[0]; err.Error() != "error getting messages: could not retrieve messages: forced error" {
			t.Logf("incorrect error: %s", err.Error())
			return false
		}

		return true
	}, createHistoryStreamTestDelay, 10*time.Millisecond)

	require.NotEmpty(t, a.FeedRepository.GetMessagesCalls())
	require.Empty(t, rw.WrittenMessages())
}

func TestCreateHistoryStream_ArgumentsAreCorrectlyPassedToRepository(t *testing.T) {
	t.Parallel()

	feed := fixtures.SomeRefFeed()
	seq := fixtures.SomeSequence()
	limit := 10

	testCases := []struct {
		Name          string
		Seq           *message.Sequence
		Limit         *int
		ExpectedCalls []mocks.FeedRepositoryMockGetMessagesCall
	}{
		{
			Name:  "nil",
			Seq:   nil,
			Limit: nil,
			ExpectedCalls: []mocks.FeedRepositoryMockGetMessagesCall{
				{
					Id:    feed,
					Seq:   nil,
					Limit: nil,
				},
			},
		},
		{
			Name:  "values",
			Seq:   &seq,
			Limit: &limit,
			ExpectedCalls: []mocks.FeedRepositoryMockGetMessagesCall{
				{
					Id:    feed,
					Seq:   &seq,
					Limit: &limit,
				},
			},
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			ctx := fixtures.TestContext(t)
			a := makeQueriesAndRunCreateHistoryStreamHandler(t, ctx)

			expectedMessages := []message.Message{
				fixtures.SomeMessage(message.MustNewSequence(1), feed),
				fixtures.SomeMessage(message.MustNewSequence(2), feed),
				fixtures.SomeMessage(message.MustNewSequence(3), feed),
			}

			a.FeedRepository.GetMessagesReturnValue = expectedMessages

			rw := newCreateHistoryStreamResponseWriterMock()

			query := queries.CreateHistoryStream{
				Id:             feed,
				Seq:            testCase.Seq,
				Limit:          testCase.Limit,
				Live:           false,
				Old:            true,
				ResponseWriter: rw,
			}

			a.Queries.CreateHistoryStream.Handle(ctx, query)

			require.Eventually(t, func() bool {
				return assert.ObjectsAreEqual(testCase.ExpectedCalls, a.FeedRepository.GetMessagesCalls())
			}, createHistoryStreamTestDelay, 10*time.Millisecond)
		})
	}
}

func TestCreateHistoryStream_OldMessagesReturnedByRepositoryAreCorrectlySentAndStreamIsClosed(t *testing.T) {
	t.Parallel()

	ctx := fixtures.TestContext(t)
	a := makeQueriesAndRunCreateHistoryStreamHandler(t, ctx)

	feed := fixtures.SomeRefFeed()

	expectedMessages := []message.Message{
		fixtures.SomeMessage(message.MustNewSequence(1), feed),
		fixtures.SomeMessage(message.MustNewSequence(2), feed),
		fixtures.SomeMessage(message.MustNewSequence(3), feed),
	}

	a.FeedRepository.GetMessagesReturnValue = expectedMessages

	rw := newCreateHistoryStreamResponseWriterMock()

	query := queries.CreateHistoryStream{
		Id:             feed,
		Seq:            internal.Ptr(fixtures.SomeSequence()),
		Limit:          internal.Ptr(int(fixtures.SomePositiveInt32())),
		Live:           false,
		Old:            true,
		ResponseWriter: rw,
	}

	a.Queries.CreateHistoryStream.Handle(ctx, query)

	require.Eventually(t, func() bool {
		return assert.ObjectsAreEqual([]error{nil}, rw.WrittenErrors())
	}, createHistoryStreamTestDelay, 10*time.Millisecond)

	require.Eventually(t, func() bool {
		return assert.ObjectsAreEqual(expectedMessages, rw.WrittenMessages())
	}, createHistoryStreamTestDelay, 10*time.Millisecond)
}

func TestCreateHistoryStream_MessagesAreNotRepeatedIfTheyWereAlreadySent(t *testing.T) {
	t.Parallel()

	feed := fixtures.SomeRefFeed()

	msg1 := fixtures.SomeMessage(message.MustNewSequence(1), feed)
	msg2 := fixtures.SomeMessage(message.MustNewSequence(2), feed)
	msg3 := fixtures.SomeMessage(message.MustNewSequence(3), feed)

	testCases := []struct {
		Name string

		Seq   *message.Sequence
		Limit *int

		Repository []message.Message
		PubSub     []message.Message

		ExpectedMessages []message.Message
		ExpectedErrors   []error
	}{
		{
			Name:  "only_repository",
			Seq:   nil,
			Limit: nil,
			Repository: []message.Message{
				msg1,
				msg2,
			},
			PubSub: nil,
			ExpectedMessages: []message.Message{
				msg1,
				msg2,
			},
			ExpectedErrors: nil,
		},
		{
			Name:       "only_pubsub",
			Seq:        nil,
			Limit:      nil,
			Repository: nil,
			PubSub: []message.Message{
				msg1,
				msg2,
			},
			ExpectedMessages: []message.Message{
				msg1,
				msg2,
			},
			ExpectedErrors: nil,
		},
		{
			Name:  "pubsub_and_repository",
			Seq:   nil,
			Limit: nil,
			Repository: []message.Message{
				msg1,
			},
			PubSub: []message.Message{
				msg2,
			},
			ExpectedMessages: []message.Message{
				msg1,
				msg2,
			},
			ExpectedErrors: nil,
		},
		{
			Name:  "pubsub_and_repository_with_overlapping_messages",
			Seq:   nil,
			Limit: nil,
			Repository: []message.Message{
				msg1,
				msg2,
			},
			PubSub: []message.Message{
				msg2,
				msg3,
			},
			ExpectedMessages: []message.Message{
				msg1,
				msg2,
				msg3,
			},
			ExpectedErrors: nil,
		},
		{
			Name:  "repository_should_enforce_sequence_by_itself",
			Seq:   internal.Ptr(message.MustNewSequence(2)),
			Limit: nil,
			Repository: []message.Message{
				msg1,
				msg2,
				msg3,
			},
			PubSub: nil,
			ExpectedMessages: []message.Message{
				msg1,
				msg2,
				msg3,
			},
			ExpectedErrors: nil,
		},
		{
			Name:  "earlier_messages_from_pubsub_should_be_omitted_if_repository_returned_something",
			Seq:   internal.Ptr(message.MustNewSequence(1)),
			Limit: nil,
			Repository: []message.Message{
				msg2,
			},
			PubSub: []message.Message{
				msg1,
				msg2,
				msg3,
			},
			ExpectedMessages: []message.Message{
				msg2,
				msg3,
			},
			ExpectedErrors: nil,
		},
		{
			Name:       "earlier_messages_from_pubsub_should_be_omitted_if_repository_returned_nothing",
			Seq:        internal.Ptr(message.MustNewSequence(1)),
			Limit:      nil,
			Repository: nil,
			PubSub: []message.Message{
				msg1,
				msg2,
				msg3,
			},
			ExpectedMessages: []message.Message{
				msg2,
				msg3,
			},
			ExpectedErrors: nil,
		},
		{
			Name:  "repository_should_enforce_limit_by_itself",
			Limit: internal.Ptr(2),
			Repository: []message.Message{
				msg1,
				msg2,
				msg3,
			},
			PubSub: nil,
			ExpectedMessages: []message.Message{
				msg1,
				msg2,
				msg3,
			},
			ExpectedErrors: []error{nil},
		},
		{
			Name:  "if_repository_exhausted_the_limit_then_live_should_do_nothing",
			Limit: internal.Ptr(2),
			Repository: []message.Message{
				msg1,
				msg2,
			},
			PubSub: []message.Message{
				msg3,
			},
			ExpectedMessages: []message.Message{
				msg1,
				msg2,
			},
			ExpectedErrors: []error{nil},
		},
		{
			Name:  "repository_should_count_towards_the_limit",
			Limit: internal.Ptr(2),
			Repository: []message.Message{
				msg1,
			},
			PubSub: []message.Message{
				msg2,
				msg3,
			},
			ExpectedMessages: []message.Message{
				msg1,
				msg2,
			},
			ExpectedErrors: []error{nil},
		},
		{
			Name:       "live_should_keep_track_of_the_limit",
			Limit:      internal.Ptr(2),
			Repository: nil,
			PubSub: []message.Message{
				msg1,
				msg2,
				msg3,
			},
			ExpectedMessages: []message.Message{
				msg1,
				msg2,
			},
			ExpectedErrors: []error{nil},
		},
		{
			Name:  "if_repository_returns_messages_out_of_order_we_do_not_check_it",
			Seq:   nil,
			Limit: nil,
			Repository: []message.Message{
				msg3,
				msg2,
				msg1,
			},
			PubSub: nil,
			ExpectedMessages: []message.Message{
				msg3,
				msg2,
				msg1,
			},
			ExpectedErrors: nil,
		},
		{
			Name:       "pubsub_is_not_required_to_handle_messages_out_of_order",
			Seq:        nil,
			Limit:      nil,
			Repository: nil,
			PubSub: []message.Message{
				msg3,
				msg2,
				msg1,
			},
			ExpectedMessages: []message.Message{
				msg3,
			},
			ExpectedErrors: nil,
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			ctx := fixtures.TestContext(t)
			a := makeQueriesAndRunCreateHistoryStreamHandler(t, ctx)

			rw := newCreateHistoryStreamResponseWriterMock()

			a.FeedRepository.GetMessagesReturnValue = testCase.Repository

			query := queries.CreateHistoryStream{
				Id:             feed,
				Seq:            testCase.Seq,
				Limit:          testCase.Limit,
				Live:           true,
				Old:            true,
				ResponseWriter: rw,
			}

			a.Queries.CreateHistoryStream.Handle(ctx, query)

			<-time.After(createHistoryStreamTestDelay)

			for _, msg := range testCase.PubSub {
				a.MessagePubSub.PublishNewMessage(msg)
			}

			<-time.After(createHistoryStreamTestDelay)

			require.Equal(t, testCase.ExpectedMessages, rw.WrittenMessages())
			if len(testCase.ExpectedErrors) == 0 {
				require.Empty(t, rw.WrittenErrors())
			} else {
				require.Equal(t, testCase.ExpectedErrors, rw.WrittenErrors())
			}
		})
	}
}

func TestCreateHistoryStream_IfLimitIsReachedBySendingLiveMessagesNoMoreMessagesAreSentAndStreamIsClosedWithNoError(t *testing.T) {
	t.Parallel()

	ctx := fixtures.TestContext(t)
	a := makeQueriesAndRunCreateHistoryStreamHandler(t, ctx)

	feed := fixtures.SomeRefFeed()
	rw := newCreateHistoryStreamResponseWriterMock()

	numMessages := 10
	limit := 5
	require.Less(t, limit, numMessages)

	query := queries.CreateHistoryStream{
		Id:             feed,
		Seq:            nil,
		Limit:          &limit,
		Live:           true,
		Old:            false,
		ResponseWriter: rw,
	}

	a.Queries.CreateHistoryStream.Handle(ctx, query)

	<-time.After(createHistoryStreamTestDelay)

	go func() {
		for i := 1; i <= numMessages; i++ {
			msg := fixtures.SomeMessage(message.MustNewSequence(i), feed)
			a.MessagePubSub.PublishNewMessage(msg) // todo context so that we can timeout?
		}
	}()

	<-time.After(createHistoryStreamTestDelay)

	require.Len(t, rw.WrittenMessages(), limit)
	require.Equal(t, []error{nil}, rw.WrittenErrors(), "should have closed the stream")
}

func TestCreateHistoryStream_IfLimitIsReachedBySendingOldMessagesTheStreamIsClosedInsteadOfGoingIntoLiveMode(t *testing.T) {
	t.Parallel()

	ctx := fixtures.TestContext(t)
	a := makeQueriesAndRunCreateHistoryStreamHandler(t, ctx)

	feed := fixtures.SomeRefFeed()
	rw := newCreateHistoryStreamResponseWriterMock()

	limit := 2
	expectedMessages := []message.Message{
		fixtures.SomeMessage(message.MustNewSequence(1), feed),
		fixtures.SomeMessage(message.MustNewSequence(2), feed),
	}
	require.Len(t, expectedMessages, limit)
	a.FeedRepository.GetMessagesReturnValue = expectedMessages

	query := queries.CreateHistoryStream{
		Id:             feed,
		Seq:            nil,
		Limit:          &limit,
		Live:           true,
		Old:            true,
		ResponseWriter: rw,
	}

	a.Queries.CreateHistoryStream.Handle(ctx, query)

	<-time.After(createHistoryStreamTestDelay)

	require.Len(t, rw.WrittenMessages(), limit)
	require.Equal(t, []error{nil}, rw.WrittenErrors(), "should have closed the stream")
}

func TestCreateHistoryStream_ErrorInOnLiveMessageClosesTheStreamWithAnError(t *testing.T) {
	t.Parallel()

	ctx := fixtures.TestContext(t)
	a := makeQueriesAndRunCreateHistoryStreamHandler(t, ctx)

	feed := fixtures.SomeRefFeed()
	rw := newCreateHistoryStreamResponseWriterMock()

	query := queries.CreateHistoryStream{
		Id:             feed,
		Seq:            nil,
		Limit:          nil,
		Live:           true,
		Old:            false,
		ResponseWriter: rw,
	}

	a.Queries.CreateHistoryStream.Handle(ctx, query)

	<-time.After(createHistoryStreamTestDelay)

	rw.WriteMessageErrorToReturn = errors.New("forced error")
	go func() {
		for i := 1; i <= 10; i++ {
			msg := fixtures.SomeMessage(message.MustNewSequence(i), feed)
			a.MessagePubSub.PublishNewMessage(msg) // todo context so that we can timeout?
		}
	}()

	<-time.After(createHistoryStreamTestDelay)

	require.Empty(t, rw.WrittenMessages())
	require.Len(t, rw.WrittenErrors(), 1)
	require.EqualError(t, rw.WrittenErrors()[0], "failed to write message: forced error")
}

func TestLiveHistoryStreams_StreamsWhichAreClosedAreClosedAndCleanedUp(t *testing.T) {
	testCtx := fixtures.TestContext(t)

	rw1 := newCreateHistoryStreamResponseWriterMock()
	ctx1, cancel1 := context.WithCancel(testCtx)
	stream1 := queries.NewHistoryStream(ctx1, queries.CreateHistoryStream{
		Id:             fixtures.SomeRefFeed(),
		Seq:            nil,
		Limit:          nil,
		Live:           true,
		Old:            false,
		ResponseWriter: rw1,
	})

	rw2 := newCreateHistoryStreamResponseWriterMock()
	ctx2, cancel2 := context.WithCancel(testCtx)
	stream2 := queries.NewHistoryStream(ctx2, queries.CreateHistoryStream{
		Id:             fixtures.SomeRefFeed(),
		Seq:            nil,
		Limit:          nil,
		Live:           true,
		Old:            false,
		ResponseWriter: rw2,
	})

	streams := queries.NewLiveHistoryStreams(logging.NewDevNullLogger())
	streams.Add(stream1)
	streams.Add(stream2)

	require.Equal(t, 2, streams.Len())

	cancel1()
	streams.CleanupClosedStreams()

	require.Empty(t, rw1.WrittenMessages())
	require.Empty(t, rw2.WrittenMessages())
	require.Equal(t, []error{nil}, rw1.WrittenErrors())
	require.Empty(t, rw2.WrittenErrors())
	require.Equal(t, 1, streams.Len())

	cancel2()
	streams.CleanupClosedStreams()

	require.Empty(t, rw1.WrittenMessages())
	require.Empty(t, rw2.WrittenMessages())
	require.Equal(t, []error{nil}, rw1.WrittenErrors())
	require.Equal(t, []error{nil}, rw2.WrittenErrors())
	require.Equal(t, 0, streams.Len())
}

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
			select {
			case <-ctx.Done():
			default:
				t.Log("run error", err)
			}
		}
	}()
	return a
}

type createHistoryStreamResponseWriterMock struct {
	writtenMessages []message.Message
	writtenErrors   []error
	lock            sync.Mutex

	WriteMessageErrorToReturn error
}

func newCreateHistoryStreamResponseWriterMock() *createHistoryStreamResponseWriterMock {
	return &createHistoryStreamResponseWriterMock{}
}

func (c *createHistoryStreamResponseWriterMock) WriteMessage(msg message.Message) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.WriteMessageErrorToReturn != nil {
		return c.WriteMessageErrorToReturn
	}
	c.writtenMessages = append(c.writtenMessages, msg)
	return nil
}

func (c *createHistoryStreamResponseWriterMock) CloseWithError(err error) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.writtenErrors = append(c.writtenErrors, err)
	return nil
}

func (c *createHistoryStreamResponseWriterMock) WrittenMessages() []message.Message {
	c.lock.Lock()
	defer c.lock.Unlock()

	tmp := make([]message.Message, len(c.writtenMessages))
	copy(tmp, c.writtenMessages)
	return tmp
}

func (c *createHistoryStreamResponseWriterMock) WrittenErrors() []error {
	c.lock.Lock()
	defer c.lock.Unlock()

	tmp := make([]error, len(c.writtenErrors))
	copy(tmp, c.writtenErrors)
	return tmp
}
