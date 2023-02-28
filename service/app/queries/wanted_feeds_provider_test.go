package queries_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/stretchr/testify/require"
)

func TestWantedFeedsRepository_GetWantedFeedsIncludesSocialGraph(t *testing.T) {
	ts, err := di.BuildTestQueries(t)
	require.NoError(t, err)

	feed := fixtures.SomeRefFeed()
	hops := fixtures.SomeHops()

	ts.SocialGraphRepository.GetSocialGraphReturnValue = graph.NewSocialGraph(map[string]graph.Hops{
		feed.String(): hops,
	})

	feeds, err := ts.WantedFeedsProvider.GetWantedFeeds()
	require.NoError(t, err)
	require.Equal(t,
		replication.MustNewWantedFeeds(
			[]replication.Contact{
				replication.MustNewContact(
					feed,
					hops,
					replication.NewEmptyFeedState(),
				),
			},
			nil,
		),
		feeds,
	)
}

func TestWantedFeedsRepository_GetWantedFeedsIncludesFeedsWantList(t *testing.T) {
	ts, err := di.BuildTestQueries(t)
	require.NoError(t, err)

	feedRef := fixtures.SomeRefFeed()

	ts.FeedWantListRepository.ListReturnValue = []refs.Feed{
		feedRef,
	}

	feeds, err := ts.WantedFeedsProvider.GetWantedFeeds()
	require.NoError(t, err)
	require.Equal(t,
		replication.MustNewWantedFeeds(
			nil,
			[]replication.WantedFeed{
				replication.MustNewWantedFeed(
					feedRef,
					replication.NewEmptyFeedState(),
				),
			},
		),
		feeds,
	)
}

func TestWantedFeedsRepository_BannedWantedFeedsAreExcluded(t *testing.T) {
	ts, err := di.BuildTestQueries(t)
	require.NoError(t, err)

	feedRef := fixtures.SomeRefFeed()
	bannedFeedRef := fixtures.SomeRefFeed()

	ts.FeedWantListRepository.ListReturnValue = []refs.Feed{
		feedRef,
		bannedFeedRef,
	}

	ts.BanListRepository.Mock(bannedFeedRef)

	feeds, err := ts.WantedFeedsProvider.GetWantedFeeds()
	require.NoError(t, err)
	require.Equal(t,
		replication.MustNewWantedFeeds(
			nil,
			[]replication.WantedFeed{
				replication.MustNewWantedFeed(
					feedRef,
					replication.NewEmptyFeedState(),
				),
			},
		),
		feeds,
	)

	require.NoError(t, err)
}
