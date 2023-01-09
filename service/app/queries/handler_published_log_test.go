package queries_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestPublishedLog_NilStartSeqIsNotUsedToDetermineMessageSequence(t *testing.T) {
	app, err := di.BuildTestQueries(t)
	require.NoError(t, err)

	localFeed := refs.MustNewIdentityFromPublic(app.LocalIdentity).MainFeed()

	query := queries.PublishedLog{
		StartSeq: nil,
	}

	_, err = app.Queries.PublishedLog.Handle(query)
	require.NoError(t, err)

	require.Empty(t, app.ReceiveLogRepository.GetMessageCalls)
	require.Equal(
		t,
		[]mocks.FeedRepositoryMockGetMessagesCall{
			{
				Id:    localFeed,
				Seq:   nil,
				Limit: nil,
			},
		},
		app.FeedRepository.GetMessagesCalls(),
	)
}

func TestPublishedLog_NotNilStartSeqIsUsedToDetermineMessageSequence(t *testing.T) {
	app, err := di.BuildTestQueries(t)
	require.NoError(t, err)

	localFeed := refs.MustNewIdentityFromPublic(app.LocalIdentity).MainFeed()

	receiveLogSequence := fixtures.SomeReceiveLogSequence()
	sequence := fixtures.SomeSequence()
	msg := fixtures.SomeMessage(sequence, localFeed)

	query := queries.PublishedLog{
		StartSeq: internal.Ptr(receiveLogSequence),
	}

	app.ReceiveLogRepository.MockMessage(receiveLogSequence, msg)

	_, err = app.Queries.PublishedLog.Handle(query)
	require.NoError(t, err)

	require.NotEmpty(t, app.ReceiveLogRepository.GetMessageCalls)
	require.Equal(
		t,
		[]mocks.FeedRepositoryMockGetMessagesCall{
			{
				Id:    localFeed,
				Seq:   internal.Ptr(sequence),
				Limit: nil,
			},
		},
		app.FeedRepository.GetMessagesCalls(),
	)
}

func TestPublishedLog_StartSequenceMustPointToMessageFromMainLocalFeed(t *testing.T) {
	app, err := di.BuildTestQueries(t)
	require.NoError(t, err)

	seq := fixtures.SomeReceiveLogSequence()
	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	require.NotEqual(t, app.LocalIdentity, msg.Feed().Identity())

	query := queries.PublishedLog{
		StartSeq: internal.Ptr(seq),
	}

	app.ReceiveLogRepository.MockMessage(seq, msg)

	_, err = app.Queries.PublishedLog.Handle(query)
	require.EqualError(t, err, "start sequence doesn't point to a message from this feed")

	require.NotEmpty(t, app.ReceiveLogRepository.GetMessageCalls)
}

func TestPublishedLog_FirstSequenceFromTheReturnedSequencesIsUsed(t *testing.T) {
	app, err := di.BuildTestQueries(t)
	require.NoError(t, err)

	localFeed := refs.MustNewIdentityFromPublic(app.LocalIdentity).MainFeed()

	receiveLogSequence1 := common.MustNewReceiveLogSequence(5)
	receiveLogSequence2 := common.MustNewReceiveLogSequence(10)

	query := queries.PublishedLog{
		StartSeq: nil,
	}

	sequence := fixtures.SomeSequence()
	msg := fixtures.SomeMessage(sequence, localFeed)

	app.FeedRepository.GetMessagesReturnValue = []message.Message{
		msg,
	}

	app.ReceiveLogRepository.MockMessage(receiveLogSequence1, msg)
	app.ReceiveLogRepository.MockMessage(receiveLogSequence2, msg)

	msgs, err := app.Queries.PublishedLog.Handle(query)
	require.NoError(t, err)

	require.Equal(
		t,
		[]mocks.FeedRepositoryMockGetMessagesCall{
			{
				Id:    localFeed,
				Seq:   nil,
				Limit: nil,
			},
		},
		app.FeedRepository.GetMessagesCalls(),
	)

	require.Equal(t,
		[]refs.Message{
			msg.Id(),
		},
		app.ReceiveLogRepository.GetSequencesCalls,
	)

	require.Equal(t,
		[]queries.LogMessage{
			{
				Message:  msg,
				Sequence: receiveLogSequence1,
			},
		},
		msgs,
	)
}
