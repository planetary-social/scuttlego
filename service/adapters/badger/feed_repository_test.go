package badger_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/stretchr/testify/require"
)

func TestFeedRepository_GetMessageReturnsMessageWhichIsStoredInRepo(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	feedRef := fixtures.SomeRefFeed()
	sequence := message.NewFirstSequence()
	msg := fixtures.SomeMessage(sequence, feedRef)

	ts.Dependencies.BanListHasher.Mock(feedRef, fixtures.SomeBanListHash())

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.FeedRepository.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
			return feed.AppendMessage(msg)
		})
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		retrievedMsg, err := adapters.FeedRepository.GetMessage(feedRef, sequence)
		require.NoError(t, err)

		// todo returned message will not match the saved message due to the way fixtures.SomeMessage works, this should be fixed
		require.Equal(t, msg.Raw(), retrievedMsg.Raw())

		return nil
	})
	require.NoError(t, err)
}

func TestFeedRepository_GetFeed_ReturnsAppropriateErrorWhenEmpty(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		_, err := adapters.FeedRepository.GetFeed(fixtures.SomeRefFeed())
		require.ErrorIs(t, err, badger.ErrFeedNotFound)

		return nil
	})
	require.NoError(t, err)
}

func TestFeedRepository_DeleteFeed(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	feedRef := fixtures.SomeRefFeed()

	ts.Dependencies.BanListHasher.Mock(feedRef, fixtures.SomeBanListHash())

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.FeedRepository.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
			return feed.AppendMessage(fixtures.SomeMessage(message.NewFirstSequence(), feedRef))
		})
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		count, err := adapters.FeedRepository.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err = adapters.FeedRepository.DeleteFeed(feedRef)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		count, err := adapters.FeedRepository.Count()
		require.NoError(t, err)
		require.Equal(t, 0, count)

		return nil
	})
	require.NoError(t, err)
}
