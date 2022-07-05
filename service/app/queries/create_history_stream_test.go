package queries_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/stretchr/testify/require"
)

func TestCreateHistoryStream_no_old_no_live(t *testing.T) {
	a, err := di.BuildTestQueries()
	require.NoError(t, err)

	ctx := fixtures.TestContext(t)

	query := queries.CreateHistoryStream{
		Id:    fixtures.SomeRefFeed(),
		Seq:   nil,
		Limit: nil,
		Live:  false,
		Old:   false,
	}

	ch := a.Queries.CreateHistoryStream.Handle(ctx, query)

	for _ = range ch {
		t.Fatal("channel should be closed right away")
	}

	require.Empty(t, a.FeedRepository.GetMessagesCalls, "since old is not specified repository shouldn't have been called")
	require.Zero(t, a.MessagePubSub.CallsCount, "since live is not specified pubsub shouldn't have been called")
}

func TestCreateHistoryStream_if_repository_returns_error_live_messages_are_not_returned(t *testing.T) {
	a, err := di.BuildTestQueries()
	require.NoError(t, err)

	ctx := fixtures.TestContext(t)

	feed := fixtures.SomeRefFeed()

	a.FeedRepository.GetMessagesReturnErr = errors.New("forced error")

	query := queries.CreateHistoryStream{
		Id:    feed,
		Seq:   nil,
		Limit: nil,
		Live:  true,
		Old:   true,
	}

	ch := a.Queries.CreateHistoryStream.Handle(ctx, query)

	var receivedValues []queries.MessageWithErr
	for msgWithError := range ch {
		receivedValues = append(receivedValues, msgWithError)
	}

	require.NotEmpty(t, a.FeedRepository.GetMessagesCalls)
	require.Len(t, receivedValues, 1)
	require.EqualError(t, receivedValues[0].Err, "could not send messages: could not retrieve messages: forced error")
}

func TestCreateHistoryStream_repository_is_called_correctly(t *testing.T) {
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

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			a, err := di.BuildTestQueries()
			require.NoError(t, err)

			ctx := fixtures.TestContext(t)

			expectedMessages := []message.Message{
				fixtures.SomeMessage(message.MustNewSequence(1), feed),
				fixtures.SomeMessage(message.MustNewSequence(2), feed),
				fixtures.SomeMessage(message.MustNewSequence(3), feed),
			}

			a.FeedRepository.GetMessagesReturnValue = expectedMessages

			query := queries.CreateHistoryStream{
				Id:    feed,
				Seq:   testCase.Seq,
				Limit: testCase.Limit,
				Live:  false,
				Old:   true,
			}

			ch := a.Queries.CreateHistoryStream.Handle(ctx, query)

			var receivedMessages []message.Message
			for msgWithError := range ch {
				require.NoError(t, msgWithError.Err)
				receivedMessages = append(receivedMessages, msgWithError.Message)
			}

			require.Equal(t, testCase.ExpectedCalls, a.FeedRepository.GetMessagesCalls)
			require.Equal(t, expectedMessages, receivedMessages)
		})
	}
}

func TestCreateHistoryStream(t *testing.T) {
	a, err := di.BuildTestQueries()
	require.NoError(t, err)

	ctx := fixtures.TestContext(t)

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
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			a.FeedRepository.GetMessagesReturnValue = testCase.Repository
			a.MessagePubSub.NewMessagesToSend = testCase.PubSub

			query := queries.CreateHistoryStream{
				Id:    feed,
				Seq:   testCase.Seq,
				Limit: testCase.Limit,
				Live:  true,
				Old:   true,
			}

			ch := a.Queries.CreateHistoryStream.Handle(ctx, query)

			var receivedMessages []message.Message
			for msgWithError := range ch {
				if msgWithError.Err != nil {
					continue
				}

				receivedMessages = append(receivedMessages, msgWithError.Message)
			}

			require.Equal(t, testCase.ExpectedMessages, receivedMessages)
		})
	}
}
