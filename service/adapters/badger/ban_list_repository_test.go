package badger_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	badgeradapters "github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/stretchr/testify/require"
)

func TestBanListRepository_LookupMappingReturnsCorrectErrorWhenMappingDoesNotExist(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		_, err := adapters.BanList.LookupMapping(fixtures.SomeBanListHash())
		require.ErrorIs(t, err, commands.ErrBanListMappingNotFound)

		return nil
	})
	require.NoError(t, err)
}

func TestBanListRepository_CreateFeedMappingInsertsMappingsAndRemoveFeedMappingRemovesMappings(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		_, err := adapters.BanList.LookupMapping(fixtures.SomeBanListHash())
		require.ErrorIs(t, err, commands.ErrBanListMappingNotFound)

		return nil
	})
	require.NoError(t, err)

	feedRef := fixtures.SomeRefFeed()
	banListHash := fixtures.SomeBanListHash()

	err = ts.TransactionProvider.Update(func(adapters badgeradapters.TestAdapters) error {
		adapters.BanListHasher.Mock(feedRef, banListHash)

		err := adapters.BanList.CreateFeedMapping(feedRef)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		bannableRef, err := adapters.BanList.LookupMapping(banListHash)
		require.NoError(t, err)

		require.Equal(t, feedRef, bannableRef.Value())

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badgeradapters.TestAdapters) error {
		adapters.BanListHasher.Mock(feedRef, banListHash)

		err := adapters.BanList.RemoveFeedMapping(feedRef)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		_, err := adapters.BanList.LookupMapping(fixtures.SomeBanListHash())
		require.ErrorIs(t, err, commands.ErrBanListMappingNotFound)

		return nil
	})
	require.NoError(t, err)
}

func TestBanListRepository_AddInsertsHashesAndRemoveRemovesHashesFromBanList(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	banListHash := fixtures.SomeBanListHash()

	err := ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		ok, err := adapters.BanList.Contains(banListHash)
		require.NoError(t, err)
		require.False(t, ok)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badgeradapters.TestAdapters) error {
		err := adapters.BanList.Add(banListHash)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		ok, err := adapters.BanList.Contains(banListHash)
		require.NoError(t, err)
		require.True(t, ok)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badgeradapters.TestAdapters) error {
		err := adapters.BanList.Remove(banListHash)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		ok, err := adapters.BanList.Contains(banListHash)
		require.NoError(t, err)
		require.False(t, ok)

		return nil
	})
	require.NoError(t, err)
}

func TestBanListRepository_ContainsFeedCorrectlyLooksUpHashes(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	feedRef := fixtures.SomeRefFeed()
	banListHash := fixtures.SomeBanListHash()

	err := ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		adapters.BanListHasher.Mock(feedRef, banListHash)

		ok, err := adapters.BanList.ContainsFeed(feedRef)
		require.NoError(t, err)
		require.False(t, ok)

		ok, err = adapters.BanList.Contains(banListHash)
		require.NoError(t, err)
		require.False(t, ok)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badgeradapters.TestAdapters) error {
		err := adapters.BanList.Add(banListHash)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		adapters.BanListHasher.Mock(feedRef, banListHash)

		ok, err := adapters.BanList.ContainsFeed(feedRef)
		require.NoError(t, err)
		require.True(t, ok)

		ok, err = adapters.BanList.Contains(banListHash)
		require.NoError(t, err)
		require.True(t, ok)

		return nil
	})
	require.NoError(t, err)
}
