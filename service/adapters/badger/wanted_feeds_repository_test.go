package badger_test

import (
	"sort"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/stretchr/testify/require"
)

func TestWantedFeedsRepository_GetWantedFeedsReturnsOnlyOwnFeedIfDatabaseIsEmpty(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	localRef := refs.MustNewIdentityFromPublic(ts.Dependencies.LocalIdentity)

	err := ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		feeds, err := adapters.WantedFeedsRepository.GetWantedFeeds()
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

		return nil
	})
	require.NoError(t, err)
}

func TestWantedFeedsRepository_GetWantedFeedsIncludesFeedsWantList(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	feedRef := fixtures.SomeRefFeed()
	ts.Dependencies.BanListHasher.Mock(feedRef, fixtures.SomeBanListHash())

	now := time.Now()
	ts.Dependencies.CurrentTimeProvider.CurrentTime = now

	localRef := refs.MustNewIdentityFromPublic(ts.Dependencies.LocalIdentity)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.FeedWantListRepository.Add(feedRef, now.Add(fixtures.SomeDuration()))
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		feeds, err := adapters.WantedFeedsRepository.GetWantedFeeds()
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
				[]replication.WantedFeed{
					{
						Who:       feedRef,
						FeedState: replication.NewEmptyFeedState(),
					},
				},
			),
			feeds,
		)

		return nil
	})
	require.NoError(t, err)
}

func TestWantedFeedsRepository_BannedContactsAreExcluded(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	feedRef := fixtures.SomeRefIdentity()
	ts.Dependencies.BanListHasher.Mock(feedRef.MainFeed(), fixtures.SomeBanListHash())

	bannedFeedRef := fixtures.SomeRefIdentity()
	bannedFeedRefBanListHash := fixtures.SomeBanListHash()
	ts.Dependencies.BanListHasher.Mock(bannedFeedRef.MainFeed(), bannedFeedRefBanListHash)

	now := time.Now()
	ts.Dependencies.CurrentTimeProvider.CurrentTime = now

	localRef := refs.MustNewIdentityFromPublic(ts.Dependencies.LocalIdentity)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.SocialGraphRepository.UpdateContact(localRef, feedRef, func(contact *feeds.Contact) error {
			return contact.Update(
				content.MustNewContactActions(
					[]content.ContactAction{content.ContactActionFollow},
				),
			)
		})
		require.NoError(t, err)

		err = adapters.SocialGraphRepository.UpdateContact(localRef, bannedFeedRef, func(contact *feeds.Contact) error {
			return contact.Update(
				content.MustNewContactActions(
					[]content.ContactAction{content.ContactActionFollow},
				),
			)
		})
		require.NoError(t, err)

		err = adapters.BanListRepository.Add(bannedFeedRefBanListHash)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		feeds, err := adapters.WantedFeedsRepository.GetWantedFeeds()
		require.NoError(t, err)

		contacts := feeds.Contacts()
		expectedContacts := []replication.Contact{
			{
				Who:       localRef.MainFeed(),
				Hops:      graph.MustNewHops(0),
				FeedState: replication.NewEmptyFeedState(),
			},
			{
				Who:       feedRef.MainFeed(),
				Hops:      graph.MustNewHops(1),
				FeedState: replication.NewEmptyFeedState(),
			},
		}

		sort.Slice(contacts, func(i, j int) bool {
			return contacts[i].Who.String() < contacts[j].Who.String()
		})

		sort.Slice(expectedContacts, func(i, j int) bool {
			return expectedContacts[i].Who.String() < expectedContacts[j].Who.String()
		})

		require.Equal(t, expectedContacts, contacts)
		require.Nil(t, feeds.OtherFeeds())

		return nil
	})
	require.NoError(t, err)
}

func TestWantedFeedsRepository_BannedWantedFeedsAreExcluded(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	feedRef := fixtures.SomeRefFeed()
	ts.Dependencies.BanListHasher.Mock(feedRef, fixtures.SomeBanListHash())

	bannedFeedRef := fixtures.SomeRefFeed()
	bannedFeedRefBanListHash := fixtures.SomeBanListHash()
	ts.Dependencies.BanListHasher.Mock(bannedFeedRef, bannedFeedRefBanListHash)

	now := time.Now()
	ts.Dependencies.CurrentTimeProvider.CurrentTime = now

	localRef := refs.MustNewIdentityFromPublic(ts.Dependencies.LocalIdentity)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.FeedWantListRepository.Add(feedRef, now.Add(fixtures.SomeDuration()))
		require.NoError(t, err)

		err = adapters.FeedWantListRepository.Add(bannedFeedRef, now.Add(fixtures.SomeDuration()))
		require.NoError(t, err)

		err = adapters.BanListRepository.Add(bannedFeedRefBanListHash)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		feeds, err := adapters.WantedFeedsRepository.GetWantedFeeds()
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
				[]replication.WantedFeed{
					{
						Who:       feedRef,
						FeedState: replication.NewEmptyFeedState(),
					},
				},
			),
			feeds,
		)

		return nil
	})
	require.NoError(t, err)
}
