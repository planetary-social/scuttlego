package badger_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestFeedRepository_GetMessageReturnsMessageWhichIsStoredInRepo(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	feedRef := fixtures.SomeRefFeed()
	sequence := message.NewFirstSequence()
	msg := fixtures.SomeMessageWithUniqueRawMessage(sequence, feedRef)

	ts.Dependencies.BanListHasher.Mock(feedRef, fixtures.SomeBanListHash())
	ts.Dependencies.RawMessageIdentifier.Mock(msg)

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
		require.Equal(t, msg, retrievedMsg)

		return nil
	})
	require.NoError(t, err)
}

func TestFeedRepository_FeedRepositoryCorrectlyLoadsFeeds(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	feedRef := fixtures.SomeRefFeed()
	ts.Dependencies.BanListHasher.Mock(feedRef, fixtures.SomeBanListHash())

	const numMessages = 10

	var messages []message.Message

	for i := 0; i < numMessages; i++ {
		seq := message.MustNewSequence(i + 1)

		var previous *refs.Message
		if !seq.IsFirst() {
			previous = internal.Ptr(messages[i-1].Id())
		}

		rawMessage := message.MustNewRawMessage(fixtures.SomeBytes())

		msg := message.MustNewMessage(
			fixtures.SomeRefMessage(),
			previous,
			seq,
			refs.MustNewIdentity(feedRef.String()),
			feedRef,
			fixtures.SomeTime(),
			fixtures.SomeContent(),
			rawMessage,
		)
		messages = append(messages, msg)

		ts.Dependencies.RawMessageIdentifier.Mock(msg)
	}

	for _, msg := range messages {
		err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
			return adapters.FeedRepository.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
				return feed.AppendMessage(msg)
			})
		})
		require.NoError(t, err, "repository should have loaded last message and created a feed so that the new message can be appended")
	}

	err := ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		feed, err := adapters.FeedRepository.GetFeed(feedRef)
		require.NoError(t, err)

		sequence, ok := feed.Sequence()
		require.True(t, ok)
		require.Equal(t, numMessages, sequence.Int())

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

func TestFeedRepository_CountUpdatesWhenUpdatingAndDeletingFeeds(t *testing.T) {
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

func TestFeedRepository_CountUpdatesOnlyWhenFirstMessageIsInserted(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	feedRef := fixtures.SomeRefFeed()

	ts.Dependencies.BanListHasher.Mock(feedRef, fixtures.SomeBanListHash())

	msg1 := fixtures.SomeMessage(message.MustNewSequence(1), feedRef)
	msg2 := fixtures.SomeMessage(message.MustNewSequence(2), feedRef)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.FeedRepository.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
			return feed.AppendMessage(msg1)
		})
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
		return adapters.FeedRepository.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
			return feed.AppendMessage(msg2)
		})
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		count, err := adapters.FeedRepository.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count)

		return nil
	})
	require.NoError(t, err)
}

func TestFeedRepository_CountDoesNotUpdateIfFeedDoesNotExist(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	feedRef := fixtures.SomeRefFeed()
	ts.Dependencies.BanListHasher.Mock(feedRef, fixtures.SomeBanListHash())

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.FeedRepository.DeleteFeed(feedRef)
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
