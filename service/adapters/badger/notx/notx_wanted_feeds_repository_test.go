package notx_test

import (
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
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

func TestNoTxWantedFeedsRepository_GetWantedFeedsCanTriggerWrites(t *testing.T) {
	ts := di.BuildBadgerNoTxTestAdapters(t)

	until := time.Now()
	afterUntil := until.Add(fixtures.SomeDuration())
	beforeUntil := until.Add(-fixtures.SomeDuration())

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		feed := fixtures.SomeRefFeed()

		err := adapters.FeedWantListRepository.Add(feed, until)
		require.NoError(t, err)

		ts.Dependencies.BanListHasher.Mock(feed, fixtures.SomeBanListHash())

		return nil
	})
	require.NoError(t, err)

	ts.Dependencies.CurrentTimeProvider.CurrentTime = beforeUntil

	l, err := ts.NoTxTestAdapters.NoTxWantedFeedsRepository.GetWantedFeeds()
	require.NoError(t, err)
	require.NotEmpty(t, l.OtherFeeds(), "if the deadline hasn't passed the value should be returned")

	ts.Dependencies.CurrentTimeProvider.CurrentTime = afterUntil

	l, err = ts.NoTxTestAdapters.NoTxWantedFeedsRepository.GetWantedFeeds()
	require.NoError(t, err)
	require.Empty(t, l.OtherFeeds(), "if the deadline passed the value shouldn't be returned")

	ts.Dependencies.CurrentTimeProvider.CurrentTime = beforeUntil

	l, err = ts.NoTxTestAdapters.NoTxWantedFeedsRepository.GetWantedFeeds()
	require.NoError(t, err)
	require.Empty(t, l.OtherFeeds(), "calling list should have cleaned up values for which the deadline has passed")
}
