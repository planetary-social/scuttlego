package bolt_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/cmd/ssb-test/di"
	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/domain/feeds"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
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

		return txadapters.FeedRepository.UpdateFeed(feedRef, func(feed *feeds.Feed) (*feeds.Feed, error) {
			if err := feed.AppendMessage(msg); err != nil {
				return nil, errors.Wrap(err, "failed to append message")
			}

			return feed, nil
		})
	})
	require.NoError(t, err)

	count, err = adapters.FeedRepository.Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
