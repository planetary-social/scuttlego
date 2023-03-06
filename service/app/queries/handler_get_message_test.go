package queries_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/internal/mocks"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/stretchr/testify/require"
)

func TestGetMessageHandler(t *testing.T) {
	tq, err := di.BuildTestQueries(t)
	require.NoError(t, err)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	query, err := queries.NewGetMessage(msg.Id())
	require.NoError(t, err)

	tq.MessageRepository.MockGet(msg)

	retrievedMsg, err := tq.Queries.GetMessage.Handle(query)
	require.NoError(t, err)
	require.Equal(t, msg, retrievedMsg)
	require.Equal(t,
		[]mocks.MessageRepositoryMockGetCall{
			{
				Id: msg.Id(),
			},
		},
		tq.MessageRepository.GetCalls(),
	)
}
