package queries_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/internal/mocks"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/di"
	"github.com/stretchr/testify/require"
)

func TestGetMessageBySequenceHandler(t *testing.T) {
	tq, err := di.BuildTestQueries(t)
	require.NoError(t, err)

	feed := fixtures.SomeRefFeed()
	sequence := fixtures.SomeSequence()

	query, err := queries.NewGetMessageBySequence(feed, sequence)
	require.NoError(t, err)

	expectedMessage := fixtures.SomeMessage(sequence, feed)
	tq.FeedRepository.MockGetMessage(expectedMessage)

	msg, err := tq.Queries.GetMessageBySequence.Handle(query)
	require.NoError(t, err)
	require.Equal(t, expectedMessage, msg)
	require.Equal(t,
		[]mocks.FeedRepositoryMockGetMessageCall{
			{
				Feed: feed,
				Seq:  sequence,
			},
		},
		tq.FeedRepository.GetMessageCalls(),
	)
}
