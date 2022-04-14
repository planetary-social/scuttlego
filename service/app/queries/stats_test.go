package queries_test

import (
	"testing"

	"github.com/planetary-social/go-ssb/cmd/ssb-test/di"
	"github.com/stretchr/testify/require"
)

func TestStats(t *testing.T) {
	a, err := di.BuildApplicationForTests()
	require.NoError(t, err)

	expectedMessageCount := 123
	expectedFeedCount := 456

	a.MessageRepository.CountReturnValue = expectedMessageCount
	a.FeedRepository.CountReturnValue = expectedFeedCount

	result, err := a.Queries.Stats.Handle()
	require.NoError(t, err)

	require.Equal(t, expectedMessageCount, result.NumberOfMessages)
	require.Equal(t, expectedFeedCount, result.NumberOfFeeds)
}
