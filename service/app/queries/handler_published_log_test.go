package queries_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestPublishedLog_IfNoSequenceIsGivenThenAllMessagesAreReturned(t *testing.T) {
	app, err := di.BuildTestQueries(t)
	require.NoError(t, err)

	localFeed := refs.MustNewIdentityFromPublic(app.LocalIdentity).MainFeed()
	msgs := mockFeedMessages(localFeed, 2)

	f := feeds.NewFeed(nil)
	for _, msg := range msgs {
		err := f.AppendMessage(msg)
		require.NoError(t, err)
	}
	app.FeedRepository.GetFeedReturnValue = f

	for i, msg := range msgs {
		app.ReceiveLogRepository.MockMessage(common.MustNewReceiveLogSequence(i), msg)
		app.FeedRepository.MockGetMessage(msg)
	}

	query := queries.PublishedLog{
		LastSeq: nil,
	}

	result, err := app.Queries.PublishedLog.Handle(query)
	require.NoError(t, err)

	require.Equal(t,
		[]queries.LogMessage{
			{
				Message:  msgs[0],
				Sequence: common.MustNewReceiveLogSequence(0),
			},
			{
				Message:  msgs[1],
				Sequence: common.MustNewReceiveLogSequence(1),
			},
		},
		result,
	)

	require.Equal(t,
		[]refs.Feed{
			localFeed,
		},
		app.FeedRepository.GetFeedCalls,
	)

	require.Equal(t,
		[]mocks.FeedRepositoryMockGetMessageCall{
			{
				Feed: localFeed,
				Seq:  msgs[1].Sequence(),
			},
			{
				Feed: localFeed,
				Seq:  msgs[0].Sequence(),
			},
		},
		app.FeedRepository.GetMessageCalls(),
	)

	require.Equal(t,
		[]refs.Message{
			msgs[1].Id(),
			msgs[0].Id(),
		},
		app.ReceiveLogRepository.GetSequencesCalls,
	)
}

func TestPublishedLog_IfSomeSequenceIsGivenThenMessagesWithHigherHighestReceiveLogSequencesAreReturned(t *testing.T) {
	app, err := di.BuildTestQueries(t)
	require.NoError(t, err)

	localFeed := refs.MustNewIdentityFromPublic(app.LocalIdentity).MainFeed()
	msgs := mockFeedMessages(localFeed, 4)

	f := feeds.NewFeed(nil)
	for _, msg := range msgs {
		err := f.AppendMessage(msg)
		require.NoError(t, err)
	}
	app.FeedRepository.GetFeedReturnValue = f

	for i, msg := range msgs {
		app.ReceiveLogRepository.MockMessage(common.MustNewReceiveLogSequence(i), msg)
		app.FeedRepository.MockGetMessage(msg)
	}

	query := queries.PublishedLog{
		LastSeq: internal.Ptr(common.MustNewReceiveLogSequence(1)),
	}

	result, err := app.Queries.PublishedLog.Handle(query)
	require.NoError(t, err)

	require.Equal(t,
		[]queries.LogMessage{
			{
				Message:  msgs[2],
				Sequence: common.MustNewReceiveLogSequence(2),
			},
			{
				Message:  msgs[3],
				Sequence: common.MustNewReceiveLogSequence(3),
			},
		},
		result,
	)

	require.Equal(t,
		[]refs.Feed{
			localFeed,
		},
		app.FeedRepository.GetFeedCalls,
	)

	require.Equal(t,
		[]mocks.FeedRepositoryMockGetMessageCall{
			{
				Feed: localFeed,
				Seq:  msgs[3].Sequence(),
			},
			{
				Feed: localFeed,
				Seq:  msgs[2].Sequence(),
			},
			{
				Feed: localFeed,
				Seq:  msgs[1].Sequence(),
			},
		},
		app.FeedRepository.GetMessageCalls(),
	)

	require.Equal(t,
		[]refs.Message{
			msgs[3].Id(),
			msgs[2].Id(),
			msgs[1].Id(),
		},
		app.ReceiveLogRepository.GetSequencesCalls,
	)
}

func TestPublishedLog_HighestSequenceFromTheReturnedSequencesIsUsed(t *testing.T) {
	app, err := di.BuildTestQueries(t)
	require.NoError(t, err)

	localFeed := refs.MustNewIdentityFromPublic(app.LocalIdentity).MainFeed()
	msgs := mockFeedMessages(localFeed, 1)

	f := feeds.NewFeed(nil)
	for _, msg := range msgs {
		err := f.AppendMessage(msg)
		require.NoError(t, err)
	}
	app.FeedRepository.GetFeedReturnValue = f

	for i, msg := range msgs {
		app.ReceiveLogRepository.MockMessage(common.MustNewReceiveLogSequence(i), msg)
		app.FeedRepository.MockGetMessage(msg)
	}

	receiveLogSequence1 := common.MustNewReceiveLogSequence(123)
	receiveLogSequence2 := common.MustNewReceiveLogSequence(345)
	receiveLogSequence3 := common.MustNewReceiveLogSequence(234)

	require.Greater(t, receiveLogSequence2.Int(), receiveLogSequence3.Int())
	require.Greater(t, receiveLogSequence2.Int(), receiveLogSequence1.Int())

	query := queries.PublishedLog{
		LastSeq: nil,
	}

	app.ReceiveLogRepository.MockMessage(receiveLogSequence1, msgs[0])
	app.ReceiveLogRepository.MockMessage(receiveLogSequence2, msgs[0])
	app.ReceiveLogRepository.MockMessage(receiveLogSequence3, msgs[0])

	result, err := app.Queries.PublishedLog.Handle(query)
	require.NoError(t, err)

	require.Equal(t,
		[]queries.LogMessage{
			{
				Message:  msgs[0],
				Sequence: receiveLogSequence2,
			},
		},
		result,
	)

	require.Equal(t,
		[]refs.Feed{
			localFeed,
		},
		app.FeedRepository.GetFeedCalls,
	)

	require.Equal(t,
		[]mocks.FeedRepositoryMockGetMessageCall{
			{
				Feed: localFeed,
				Seq:  msgs[0].Sequence(),
			},
		},
		app.FeedRepository.GetMessageCalls(),
	)

	require.Equal(t,
		[]refs.Message{
			msgs[0].Id(),
		},
		app.ReceiveLogRepository.GetSequencesCalls,
	)
}

func mockFeedMessages(feed refs.Feed, numberOfMessages int) []message.Message {
	var messages []message.Message
	for i := 0; i < numberOfMessages; i++ {
		seq := message.MustNewSequence(i + 1)

		var previous *refs.Message
		if !seq.IsFirst() {
			previous = internal.Ptr(messages[i-1].Id())
		}

		messages = append(messages, message.MustNewMessage(
			fixtures.SomeRefMessage(),
			previous,
			seq,
			refs.MustNewIdentityFromPublic(feed.Identity()),
			feed,
			fixtures.SomeTime(),
			fixtures.SomeContent(),
			fixtures.SomeRawMessage(),
		))
	}
	return messages
}
