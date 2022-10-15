package queries_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	a, err := di.BuildTestQueries(t)
	require.NoError(t, err)

	ctx := fixtures.TestContext(t)

	expectedMessageCount := 123
	expectedFeedCount := 456

	remote := fixtures.SomePublicIdentity()

	a.MessageRepository.CountReturnValue = expectedMessageCount
	a.FeedRepository.CountReturnValue = expectedFeedCount
	a.PeerManager.MockPeers([]transport.Peer{
		transport.MustNewPeer(remote, mocks.NewConnectionMock(ctx)),
	})

	result, err := a.Queries.Status.Handle()
	require.NoError(t, err)

	require.Equal(t, expectedMessageCount, result.NumberOfMessages)
	require.Equal(t, expectedFeedCount, result.NumberOfFeeds)
	require.Equal(t,
		[]queries.Peer{
			{
				Identity: remote,
			},
		},
		result.Peers,
	)
}
