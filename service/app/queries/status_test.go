package queries_test

import (
	"testing"

	"github.com/planetary-social/go-ssb/di"
	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/app/queries"
	"github.com/planetary-social/go-ssb/service/domain/transport"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	a, err := di.BuildTestQueries()
	require.NoError(t, err)

	expectedMessageCount := 123
	expectedFeedCount := 456

	remote := fixtures.SomePublicIdentity()

	a.MessageRepository.CountReturnValue = expectedMessageCount
	a.FeedRepository.CountReturnValue = expectedFeedCount
	a.PeerManager.PeersReturnValue = []transport.Peer{
		transport.NewPeer(remote, nil),
	}

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
