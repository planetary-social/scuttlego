package notx_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/stretchr/testify/require"
)

func TestNoTxFeedRepository_GetMessage(t *testing.T) {
	ts := di.BuildBadgerNoTxTestAdapters(t)

	feedRef := fixtures.SomeRefFeed()
	sequence := message.NewFirstSequence()
	msg := fixtures.SomeMessageWithUniqueRawMessage(sequence, feedRef)

	ts.Dependencies.BanListHasher.Mock(feedRef, fixtures.SomeBanListHash())
	ts.Dependencies.RawMessageIdentifier.Mock(msg)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.FeedRepository.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
			return feed.AppendMessage(msg)
		})
	})
	require.NoError(t, err)

	retrievedMsg, err := ts.NoTxTestAdapters.NoTxFeedRepository.GetMessage(feedRef, sequence)
	require.NoError(t, err)
	require.Equal(t, msg, retrievedMsg)
}

func TestNoTxFeedRepository_Count(t *testing.T) {
	ts := di.BuildBadgerNoTxTestAdapters(t)

	count, err := ts.NoTxTestAdapters.NoTxFeedRepository.Count()
	require.NoError(t, err)
	require.Equal(t, 0, count)

	feedRef := fixtures.SomeRefFeed()
	ts.Dependencies.BanListHasher.Mock(feedRef, fixtures.SomeBanListHash())

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		msg := fixtures.SomeMessage(message.NewFirstSequence(), feedRef)

		return adapters.FeedRepository.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
			return feed.AppendMessage(msg)
		})
	})
	require.NoError(t, err)

	count, err = ts.NoTxTestAdapters.NoTxFeedRepository.Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestNoTxFeedRepository_GetMessages(t *testing.T) {
	ts := di.BuildBadgerNoTxTestAdapters(t)

	feedRef := fixtures.SomeRefFeed()
	ts.Dependencies.BanListHasher.Mock(feedRef, fixtures.SomeBanListHash())

	msgs, err := ts.NoTxTestAdapters.NoTxFeedRepository.GetMessages(feedRef, nil, nil)
	require.NoError(t, err)
	require.Empty(t, msgs)

	msg := fixtures.SomeMessage(message.NewFirstSequence(), feedRef)
	ts.Dependencies.RawMessageIdentifier.Mock(msg)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.FeedRepository.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
			return feed.AppendMessage(msg)
		})
	})
	require.NoError(t, err)

	msgs, err = ts.NoTxTestAdapters.NoTxFeedRepository.GetMessages(feedRef, nil, nil)
	require.NoError(t, err)
	require.Equal(t, []message.Message{msg}, msgs)
}
