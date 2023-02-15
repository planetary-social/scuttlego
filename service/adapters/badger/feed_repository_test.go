package badger_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/app/common"
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

func TestFeedRepository_GetMessageReturnsCorrectErrorIfMessageCanNotBeFound(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		_, err := adapters.FeedRepository.GetMessage(fixtures.SomeRefFeed(), fixtures.SomeSequence())
		require.ErrorIs(t, err, common.ErrFeedMessageNotFound)

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

func TestFeedRepository_GetFeedMessagesReturnsAllMessages(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	feedRef := fixtures.SomeRefFeed()
	authorRef := refs.MustNewIdentity(feedRef.String())

	msg1RawMessage := message.MustNewRawMessage(fixtures.SomeBytes())
	msg1 := message.MustNewMessage(
		fixtures.SomeRefMessage(),
		nil,
		message.MustNewSequence(1),
		authorRef,
		feedRef,
		fixtures.SomeTime(),
		fixtures.SomeContent(),
		msg1RawMessage,
	)

	msg2RawMessage := message.MustNewRawMessage(fixtures.SomeBytes())
	msg2 := message.MustNewMessage(
		fixtures.SomeRefMessage(),
		internal.Ptr(msg1.Id()),
		message.MustNewSequence(2),
		authorRef,
		feedRef,
		fixtures.SomeTime(),
		fixtures.SomeContent(),
		msg2RawMessage,
	)

	ts.Dependencies.BanListHasher.Mock(feedRef, fixtures.SomeBanListHash())

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.FeedRepository.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
			err := feed.AppendMessage(msg1)
			require.NoError(t, err)

			err = feed.AppendMessage(msg2)
			require.NoError(t, err)

			return nil
		})
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		msgs, err := adapters.FeedRepository.GetFeedMessages(feedRef)
		require.NoError(t, err)

		require.Equal(t,
			[]badger.FeedMessage{
				{
					Sequence: msg1.Sequence(),
					Id:       msg1.Id(),
				},
				{
					Sequence: msg2.Sequence(),
					Id:       msg2.Id(),
				},
			},
			msgs,
		)

		return nil
	})
	require.NoError(t, err)
}

func TestFeedRepository_GetFeed_ReturnsAppropriateErrorWhenEmpty(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		_, err := adapters.FeedRepository.GetFeed(fixtures.SomeRefFeed())
		require.ErrorIs(t, err, common.ErrFeedNotFound)

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

func TestFeedRepository_DeleteRemovesDataFromChildRepositories(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	feedRef := fixtures.SomeRefFeed()
	authorRef := refs.MustNewIdentityFromPublic(feedRef.Identity())
	banListHash := fixtures.SomeBanListHash()

	msg1 := message.MustNewMessage(
		fixtures.SomeRefMessage(),
		nil,
		message.MustNewSequence(1),
		authorRef,
		feedRef,
		fixtures.SomeTime(),
		fixtures.SomeContent(),
		message.MustNewRawMessage(fixtures.SomeBytes()),
	)

	msg2 := message.MustNewMessage(
		fixtures.SomeRefMessage(),
		internal.Ptr(msg1.Id()),
		message.MustNewSequence(2),
		authorRef,
		feedRef,
		fixtures.SomeTime(),
		fixtures.SomeContent(),
		message.MustNewRawMessage(fixtures.SomeBytes()),
	)

	msgs := []message.Message{
		msg1,
		msg2,
	}

	ts.Dependencies.BanListHasher.Mock(feedRef, banListHash)
	for _, msg := range msgs {
		ts.Dependencies.RawMessageIdentifier.Mock(msg)
	}

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.FeedRepository.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
			for _, msg := range msgs {
				if err := feed.AppendMessage(msg); err != nil {
					return errors.Wrap(err, "append error")
				}
			}
			return nil
		})
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		_, err := adapters.FeedRepository.GetFeed(feedRef)
		require.NoError(t, err)

		_, err = adapters.BanListRepository.LookupMapping(banListHash)
		require.NoError(t, err)

		for i, msg := range msgs {
			retrievedMsg, err := adapters.MessageRepository.Get(msg.Id())
			require.NoError(t, err)
			require.Equal(t, msg, retrievedMsg)

			_, err = adapters.ReceiveLogRepository.GetSequences(msg.Id())
			require.NoError(t, err)

			_, err = adapters.ReceiveLogRepository.GetMessage(common.MustNewReceiveLogSequence(i))
			require.NoError(t, err)
		}

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.FeedRepository.DeleteFeed(feedRef)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		_, err := adapters.FeedRepository.GetFeed(feedRef)
		require.EqualError(t, err, "feed not found")

		_, err = adapters.BanListRepository.LookupMapping(banListHash)
		require.EqualError(t, err, "ban list mapping not found")

		for i, msg := range msgs {
			_, err = adapters.MessageRepository.Get(msg.Id())
			require.EqualError(t, err, "message not found: Key not found")

			_, err = adapters.ReceiveLogRepository.GetSequences(msg.Id())
			require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)

			_, err = adapters.ReceiveLogRepository.GetMessage(common.MustNewReceiveLogSequence(i))
			require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)
		}

		return nil
	})
	require.NoError(t, err)
}

func TestFeedRepository_RemoveMessagesAtOrAboveSequenceCorrectlyRemovesMessages(t *testing.T) {
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
		require.Equal(t, message.MustNewSequence(numMessages), sequence)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.FeedRepository.RemoveMessagesAtOrAboveSequence(feedRef, message.MustNewSequence(5))
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		feed, err := adapters.FeedRepository.GetFeed(feedRef)
		require.NoError(t, err)

		sequence, ok := feed.Sequence()
		require.True(t, ok)
		require.Equal(t, message.MustNewSequence(4), sequence)

		return nil
	})
	require.NoError(t, err)
}
