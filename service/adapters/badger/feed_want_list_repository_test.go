package badger_test

import (
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/stretchr/testify/require"
)

func TestFeedWantListRepository_ListDoesNotReturnValuesForWhichUntilIsBeforeCurrentTime(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		until := time.Now()
		afterUntil := until.Add(fixtures.SomeDuration())
		beforeUntil := until.Add(-fixtures.SomeDuration())

		err := adapters.FeedWantList.Add(fixtures.SomeRefFeed(), until)
		require.NoError(t, err)

		adapters.CurrentTimeProvider.CurrentTime = beforeUntil

		l, err := adapters.FeedWantList.List()
		require.NoError(t, err)
		require.NotEmpty(t, l, "if the deadline hasn't passed the value should be returned")

		adapters.CurrentTimeProvider.CurrentTime = afterUntil

		l, err = adapters.FeedWantList.List()
		require.NoError(t, err)
		require.Empty(t, l, "if the deadline passed the value shouldn't be returned")

		adapters.CurrentTimeProvider.CurrentTime = beforeUntil

		l, err = adapters.FeedWantList.List()
		require.NoError(t, err)
		require.Empty(t, l, "calling list should have cleaned up values for which the deadline has passed")

		return nil
	})
	require.NoError(t, err)
}

func TestFeedWantListRepository_LongerUntilOverwritesShorterUntil(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		firstUntil := time.Now()
		afterFirstUntil := firstUntil.Add(fixtures.SomeDuration())
		secondUntil := afterFirstUntil.Add(fixtures.SomeDuration())

		err := adapters.FeedWantList.Add(fixtures.SomeRefFeed(), firstUntil)
		require.NoError(t, err)

		err = adapters.FeedWantList.Add(fixtures.SomeRefFeed(), secondUntil)
		require.NoError(t, err)

		adapters.CurrentTimeProvider.CurrentTime = afterFirstUntil

		l, err := adapters.FeedWantList.List()
		require.NoError(t, err)
		require.NotEmpty(t, l, "if the deadline hasn't passed the value should be returned")

		return nil
	})
	require.NoError(t, err)
}

func TestFeedWantListRepository_ShorterUntilDoesNotOverwriteLongerUntil(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		firstUntil := time.Now()
		afterFirstUntil := firstUntil.Add(fixtures.SomeDuration())
		secondUntil := afterFirstUntil.Add(fixtures.SomeDuration())

		err := adapters.FeedWantList.Add(fixtures.SomeRefFeed(), secondUntil)
		require.NoError(t, err)

		err = adapters.FeedWantList.Add(fixtures.SomeRefFeed(), firstUntil)
		require.NoError(t, err)

		adapters.CurrentTimeProvider.CurrentTime = afterFirstUntil

		l, err := adapters.FeedWantList.List()
		require.NoError(t, err)
		require.NotEmpty(t, l, "if the deadline hasn't passed the value should be returned")

		return nil
	})
	require.NoError(t, err)
}

func TestFeedWantListRepository_Contains(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		until := time.Now()
		now := until.Add(-fixtures.SomeDuration())
		adapters.CurrentTimeProvider.CurrentTime = now

		id := fixtures.SomeRefFeed()

		ok, err := adapters.FeedWantList.Contains(id)
		require.NoError(t, err)
		require.False(t, ok)

		err = adapters.FeedWantList.Add(id, until)
		require.NoError(t, err)

		ok, err = adapters.FeedWantList.Contains(id)
		require.NoError(t, err)
		require.True(t, ok)

		return nil
	})
	require.NoError(t, err)
}
