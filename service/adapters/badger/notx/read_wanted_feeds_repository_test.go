package notx_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/stretchr/testify/require"
)

func TestNoTxWantedFeedsRepository_GetWantedFeeds(t *testing.T) {
	ts := di.BuildBadgerNoTxTestAdapters(t)

	localRef := refs.MustNewIdentityFromPublic(ts.Dependencies.LocalIdentity)

	feeds, err := ts.NoTxTestAdapters.NoTxWantedFeedsRepository.GetWantedFeeds()
	require.NoError(t, err)

	require.Equal(t,
		replication.NewWantedFeeds(
			[]replication.Contact{
				{
					Who:       localRef.MainFeed(),
					Hops:      graph.MustNewHops(0),
					FeedState: replication.NewEmptyFeedState(),
				},
			},
			nil,
		),
		feeds,
	)
}
