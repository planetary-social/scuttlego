package bolt_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/bolt"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestFeedRepository_GetFeed_ReturnsAppropriateErrorWhenEmpty(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		_, err = adapters.FeedRepository.GetFeed(fixtures.SomeRefFeed())
		require.ErrorIs(t, err, bolt.ErrFeedNotFound)

		return nil
	})
	require.NoError(t, err)
}

func TestFeedRepository_DeleteFeed(t *testing.T) {
	db := fixtures.Bolt(t)

	feedRef := fixtures.SomeRefFeed()

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		err = adapters.FeedRepository.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
			return feed.AppendMessage(fixtures.SomeMessage(message.NewFirstSequence(), feedRef))
		})
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = db.View(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		count, err := adapters.FeedRepository.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		err = adapters.FeedRepository.DeleteFeed(feedRef)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = db.View(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		count, err := adapters.FeedRepository.Count()
		require.NoError(t, err)
		require.Equal(t, 0, count)

		return nil
	})
	require.NoError(t, err)
}
