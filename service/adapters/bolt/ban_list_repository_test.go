package bolt_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestBanListRepository_LookupMappingReturnsCorrectErrorWhenMappingDoesNotExist(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		_, err = adapters.BanList.LookupMapping(fixtures.SomeBanListHash())
		require.ErrorIs(t, err, commands.ErrBanListMappingNotFound)

		return nil
	})
	require.NoError(t, err)
}

func TestBanListRepository_CreateFeedMappingInsertsMappingsAndRemoveFeedMappingRemovesMappings(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		_, err = adapters.BanList.LookupMapping(fixtures.SomeBanListHash())
		require.ErrorIs(t, err, commands.ErrBanListMappingNotFound)

		return nil
	})
	require.NoError(t, err)

	feedRef := fixtures.SomeRefFeed()
	banListHash := fixtures.SomeBanListHash()

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		adapters.BanListHasher.Mock(feedRef, banListHash)

		err = adapters.BanList.CreateFeedMapping(feedRef)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		bannableRef, err := adapters.BanList.LookupMapping(banListHash)
		require.NoError(t, err)

		require.Equal(t, feedRef, bannableRef.Value())

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		adapters.BanListHasher.Mock(feedRef, banListHash)

		err = adapters.BanList.RemoveFeedMapping(feedRef)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		_, err = adapters.BanList.LookupMapping(fixtures.SomeBanListHash())
		require.ErrorIs(t, err, commands.ErrBanListMappingNotFound)

		return nil
	})
	require.NoError(t, err)
}

func TestBanListRepository_AddInsertsHashesAndRemoveRemovesHashesFromBanList(t *testing.T) {
	db := fixtures.Bolt(t)

	banListHash := fixtures.SomeBanListHash()

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		ok, err := adapters.BanList.Contains(banListHash)
		require.NoError(t, err)
		require.False(t, ok)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		err = adapters.BanList.Add(banListHash)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		ok, err := adapters.BanList.Contains(banListHash)
		require.NoError(t, err)
		require.True(t, ok)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		err = adapters.BanList.Remove(banListHash)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		ok, err := adapters.BanList.Contains(banListHash)
		require.NoError(t, err)
		require.False(t, ok)

		return nil
	})
	require.NoError(t, err)
}

func TestBanListRepository_ContainsFeedCorrectlyLooksUpHashes(t *testing.T) {
	db := fixtures.Bolt(t)

	feedRef := fixtures.SomeRefFeed()
	banListHash := fixtures.SomeBanListHash()

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

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

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		err = adapters.BanList.Add(banListHash)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

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
