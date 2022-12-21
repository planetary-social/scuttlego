package badger

import (
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/stretchr/testify/require"
)

func testBanListRepository_LookupMappingReturnsCorrectErrorWhenMappingDoesNotExist(t *testing.T, ts *TestedSystem) {
	err := ts.TransactionProvider.View(func(adapters Adapters) error {
		_, err := adapters.BanList.LookupMapping(fixtures.SomeBanListHash())
		require.ErrorIs(t, err, commands.ErrBanListMappingNotFound)

		return nil
	})
	require.NoError(t, err)
}

func testBanListRepository_CreateFeedMappingInsertsMappingsAndRemoveFeedMappingRemovesMappings(t *testing.T, ts *TestedSystem) {
	err := ts.TransactionProvider.View(func(adapters Adapters) error {
		_, err := adapters.BanList.LookupMapping(fixtures.SomeBanListHash())
		require.ErrorIs(t, err, commands.ErrBanListMappingNotFound)

		return nil
	})
	require.NoError(t, err)

	feedRef := fixtures.SomeRefFeed()
	banListHash := fixtures.SomeBanListHash()

	err = ts.TransactionProvider.Update(func(adapters Adapters) error {
		adapters.BanListHasher.Mock(feedRef, banListHash)

		err := adapters.BanList.CreateFeedMapping(feedRef)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters Adapters) error {
		bannableRef, err := adapters.BanList.LookupMapping(banListHash)
		require.NoError(t, err)

		require.Equal(t, feedRef, bannableRef.Value())

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters Adapters) error {
		adapters.BanListHasher.Mock(feedRef, banListHash)

		err := adapters.BanList.RemoveFeedMapping(feedRef)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters Adapters) error {
		_, err := adapters.BanList.LookupMapping(fixtures.SomeBanListHash())
		require.ErrorIs(t, err, commands.ErrBanListMappingNotFound)

		return nil
	})
	require.NoError(t, err)
}

func testBanListRepository_AddInsertsHashesAndRemoveRemovesHashesFromBanList(t *testing.T, ts *TestedSystem) {
	banListHash := fixtures.SomeBanListHash()

	err := ts.TransactionProvider.View(func(adapters Adapters) error {
		ok, err := adapters.BanList.Contains(banListHash)
		require.NoError(t, err)
		require.False(t, ok)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters Adapters) error {
		err := adapters.BanList.Add(banListHash)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters Adapters) error {
		ok, err := adapters.BanList.Contains(banListHash)
		require.NoError(t, err)
		require.True(t, ok)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters Adapters) error {
		err := adapters.BanList.Remove(banListHash)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters Adapters) error {
		ok, err := adapters.BanList.Contains(banListHash)
		require.NoError(t, err)
		require.False(t, ok)

		return nil
	})
	require.NoError(t, err)
}

func testBanListRepository_ContainsFeedCorrectlyLooksUpHashes(t *testing.T, ts *TestedSystem) {
	feedRef := fixtures.SomeRefFeed()
	banListHash := fixtures.SomeBanListHash()

	err := ts.TransactionProvider.View(func(adapters Adapters) error {
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

	err = ts.TransactionProvider.Update(func(adapters Adapters) error {
		err := adapters.BanList.Add(banListHash)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters Adapters) error {
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
