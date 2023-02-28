package badger_test

import (
	"bytes"
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	badgeradapters "github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/bans"
	"github.com/stretchr/testify/require"
)

func TestBanListRepository_LookupMappingReturnsCorrectErrorWhenMappingDoesNotExist(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		_, err := adapters.BanListRepository.LookupMapping(fixtures.SomeBanListHash())
		require.ErrorIs(t, err, commands.ErrBanListMappingNotFound)

		return nil
	})
	require.NoError(t, err)
}

func TestBanListRepository_CreateFeedMappingInsertsMappingsAndRemoveFeedMappingRemovesMappings(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		_, err := adapters.BanListRepository.LookupMapping(fixtures.SomeBanListHash())
		require.ErrorIs(t, err, commands.ErrBanListMappingNotFound)

		return nil
	})
	require.NoError(t, err)

	feedRef := fixtures.SomeRefFeed()
	banListHash := fixtures.SomeBanListHash()

	ts.Dependencies.BanListHasher.Mock(feedRef, banListHash)

	err = ts.TransactionProvider.Update(func(adapters badgeradapters.TestAdapters) error {
		err := adapters.BanListRepository.CreateFeedMapping(feedRef)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		bannableRef, err := adapters.BanListRepository.LookupMapping(banListHash)
		require.NoError(t, err)

		require.Equal(t, feedRef, bannableRef.Value())

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badgeradapters.TestAdapters) error {
		err := adapters.BanListRepository.RemoveFeedMapping(feedRef)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		_, err := adapters.BanListRepository.LookupMapping(fixtures.SomeBanListHash())
		require.ErrorIs(t, err, commands.ErrBanListMappingNotFound)

		return nil
	})
	require.NoError(t, err)
}

func TestBanListRepository_AddInsertsHashesAndRemoveRemovesHashesFromBanList(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	banListHash := fixtures.SomeBanListHash()

	err := ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		ok, err := adapters.BanListRepository.Contains(banListHash)
		require.NoError(t, err)
		require.False(t, ok)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badgeradapters.TestAdapters) error {
		err := adapters.BanListRepository.Add(banListHash)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		ok, err := adapters.BanListRepository.Contains(banListHash)
		require.NoError(t, err)
		require.True(t, ok)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badgeradapters.TestAdapters) error {
		err := adapters.BanListRepository.Remove(banListHash)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		ok, err := adapters.BanListRepository.Contains(banListHash)
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

	ts.Dependencies.BanListHasher.Mock(feedRef, banListHash)

	err := ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		ok, err := adapters.BanListRepository.ContainsFeed(feedRef)
		require.NoError(t, err)
		require.False(t, ok)

		ok, err = adapters.BanListRepository.Contains(banListHash)
		require.NoError(t, err)
		require.False(t, ok)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badgeradapters.TestAdapters) error {
		err := adapters.BanListRepository.Add(banListHash)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		ok, err := adapters.BanListRepository.ContainsFeed(feedRef)
		require.NoError(t, err)
		require.True(t, ok)

		ok, err = adapters.BanListRepository.Contains(banListHash)
		require.NoError(t, err)
		require.True(t, ok)

		return nil
	})
	require.NoError(t, err)
}

func TestBanListRepository_ListReturnsHashesAddedToTheRepository(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	banListHash1 := fixtures.SomeBanListHash()
	banListHash2 := fixtures.SomeBanListHash()

	err := ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		hashes, err := adapters.BanListRepository.List()
		require.NoError(t, err)
		require.Empty(t, hashes)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badgeradapters.TestAdapters) error {
		err := adapters.BanListRepository.Add(banListHash1)
		require.NoError(t, err)

		err = adapters.BanListRepository.Add(banListHash2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	expectedHashes := []bans.Hash{banListHash1, banListHash2}

	err = ts.TransactionProvider.View(func(adapters badgeradapters.TestAdapters) error {
		hashes, err := adapters.BanListRepository.List()
		require.NoError(t, err)

		compare := func(a, b bans.Hash) bool {
			return bytes.Compare(a.Bytes(), b.Bytes()) < 0
		}
		internal.SortSlice(expectedHashes, compare)
		internal.SortSlice(hashes, compare)
		require.Equal(t, expectedHashes, hashes)

		return nil
	})
	require.NoError(t, err)
}
