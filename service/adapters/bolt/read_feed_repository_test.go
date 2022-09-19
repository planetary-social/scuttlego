package bolt_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestReadFeedRepository_Count(t *testing.T) {
	db := fixtures.Bolt(t)

	adapters, err := di.BuildTestAdapters(db)
	require.NoError(t, err)

	count, err := adapters.FeedRepository.Count()
	require.NoError(t, err)
	require.Equal(t, 0, count)

	err = db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		feedRef := fixtures.SomeRefFeed()
		msg := fixtures.SomeMessage(message.NewFirstSequence(), feedRef)

		return txadapters.FeedRepository.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
			return feed.AppendMessage(msg)
		})
	})
	require.NoError(t, err)

	count, err = adapters.FeedRepository.Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
