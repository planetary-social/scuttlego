package badger_test

import (
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/di"
	"github.com/stretchr/testify/require"
)

func TestFeedWantListRepository_ListDoesNotReturnValuesForWhichUntilIsBeforeCurrentTime(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		until := time.Now()
		afterUntil := until.Add(fixtures.SomeDuration())
		beforeUntil := until.Add(-fixtures.SomeDuration())

		err := adapters.FeedWantListRepository.Add(fixtures.SomeRefFeed(), until)
		require.NoError(t, err)

		ts.Dependencies.CurrentTimeProvider.CurrentTime = beforeUntil

		l, err := adapters.FeedWantListRepository.List()
		require.NoError(t, err)
		require.NotEmpty(t, l, "if the deadline hasn't passed the value should be returned")

		ts.Dependencies.CurrentTimeProvider.CurrentTime = afterUntil

		l, err = adapters.FeedWantListRepository.List()
		require.NoError(t, err)
		require.Empty(t, l, "if the deadline passed the value shouldn't be returned")

		ts.Dependencies.CurrentTimeProvider.CurrentTime = beforeUntil

		l, err = adapters.FeedWantListRepository.List()
		require.NoError(t, err)
		require.NotEmpty(t, l, "if the deadline hasn't passed the value should be returned")

		return nil
	})
	require.NoError(t, err)
}

func TestFeedWantListRepository_Cleanup(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		until := time.Now()
		afterUntil := until.Add(fixtures.SomeDuration())
		beforeUntil := until.Add(-fixtures.SomeDuration())

		err := adapters.FeedWantListRepository.Add(fixtures.SomeRefFeed(), until)
		require.NoError(t, err)

		ts.Dependencies.CurrentTimeProvider.CurrentTime = beforeUntil

		l, err := adapters.FeedWantListRepository.List()
		require.NoError(t, err)
		require.NotEmpty(t, l, "if the deadline hasn't passed the value should be returned")

		err = adapters.FeedWantListRepository.Cleanup()
		require.NoError(t, err)

		l, err = adapters.FeedWantListRepository.List()
		require.NoError(t, err)
		require.NotEmpty(t, l, "if the deadline hasn't passed cleanup shouldn't have done anything")

		ts.Dependencies.CurrentTimeProvider.CurrentTime = afterUntil

		err = adapters.FeedWantListRepository.Cleanup()
		require.NoError(t, err)

		ts.Dependencies.CurrentTimeProvider.CurrentTime = beforeUntil

		l, err = adapters.FeedWantListRepository.List()
		require.NoError(t, err)
		require.Empty(t, l, "cleanup should have worked now")

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

		err := adapters.FeedWantListRepository.Add(fixtures.SomeRefFeed(), firstUntil)
		require.NoError(t, err)

		err = adapters.FeedWantListRepository.Add(fixtures.SomeRefFeed(), secondUntil)
		require.NoError(t, err)

		ts.Dependencies.CurrentTimeProvider.CurrentTime = afterFirstUntil

		l, err := adapters.FeedWantListRepository.List()
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

		err := adapters.FeedWantListRepository.Add(fixtures.SomeRefFeed(), secondUntil)
		require.NoError(t, err)

		err = adapters.FeedWantListRepository.Add(fixtures.SomeRefFeed(), firstUntil)
		require.NoError(t, err)

		ts.Dependencies.CurrentTimeProvider.CurrentTime = afterFirstUntil

		l, err := adapters.FeedWantListRepository.List()
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
		ts.Dependencies.CurrentTimeProvider.CurrentTime = now

		id := fixtures.SomeRefFeed()

		ok, err := adapters.FeedWantListRepository.Contains(id)
		require.NoError(t, err)
		require.False(t, ok)

		err = adapters.FeedWantListRepository.Add(id, until)
		require.NoError(t, err)

		ok, err = adapters.FeedWantListRepository.Contains(id)
		require.NoError(t, err)
		require.True(t, ok)

		return nil
	})
	require.NoError(t, err)
}
