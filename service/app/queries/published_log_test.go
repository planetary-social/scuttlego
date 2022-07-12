package queries_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestPublishedLog_NilStartSeqIsNotUsedToDetermineMessageSequence(t *testing.T) {
	app, err := di.BuildTestQueries()
	require.NoError(t, err)

	query := queries.PublishedLog{
		StartSeq: nil,
	}

	_, err = app.Queries.PublishedLog.Handle(query)
	require.NoError(t, err)

	require.Empty(t, app.ReceiveLogRepository.GetMessageCalls)
}

func TestPublishedLog_NotNilStarSeqIsUsedToDetermineMessageSequence(t *testing.T) {
	app, err := di.BuildTestQueries()
	require.NoError(t, err)

	localFeed := refs.MustNewIdentityFromPublic(app.LocalIdentity).MainFeed()

	seq := fixtures.SomeReceiveLogSequence()
	msg := fixtures.SomeMessage(fixtures.SomeSequence(), localFeed)

	query := queries.PublishedLog{
		StartSeq: internal.Ptr(seq),
	}

	app.ReceiveLogRepository.MockMessage(seq, msg)

	_, err = app.Queries.PublishedLog.Handle(query)
	require.NoError(t, err)

	require.NotEmpty(t, app.ReceiveLogRepository.GetMessageCalls)
}

func TestPublishedLog_StartSequenceMustPointToMessageFromMainLocalFeed(t *testing.T) {
	app, err := di.BuildTestQueries()
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
