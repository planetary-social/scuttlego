package bolt_test

import (
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestWantListRepositoryListDoesNotReturnValuesForWhichUntilIsBeforeCurrentTime(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		until := time.Now()
		afterUntil := until.Add(fixtures.SomeDuration())
		beforeUntil := until.Add(-fixtures.SomeDuration())

		err = txadapters.WantList.AddToWantList(fixtures.SomeRefBlob(), until)
		require.NoError(t, err)

		txadapters.CurrentTimeProvider.CurrentTime = beforeUntil

		l, err := txadapters.WantList.List()
		require.NoError(t, err)
		require.NotEmpty(t, l, "if the deadline hasn't passed the value should be returned")

		txadapters.CurrentTimeProvider.CurrentTime = afterUntil

		l, err = txadapters.WantList.List()
		require.NoError(t, err)
		require.Empty(t, l, "if the deadline passed the value shouldn't be returned")

		txadapters.CurrentTimeProvider.CurrentTime = beforeUntil

		l, err = txadapters.WantList.List()
		require.NoError(t, err)
		require.Empty(t, l, "calling list should have cleaned up values for which the deadline has passed")

		return nil
	})
	require.NoError(t, err)
}

func TestWantListRepositoryLongerUntilOverwritesShorterUntil(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		firstUntil := time.Now()
		afterFirstUntil := firstUntil.Add(fixtures.SomeDuration())
		secondUntil := afterFirstUntil.Add(fixtures.SomeDuration())

		err = txadapters.WantList.AddToWantList(fixtures.SomeRefBlob(), firstUntil)
		require.NoError(t, err)

		err = txadapters.WantList.AddToWantList(fixtures.SomeRefBlob(), secondUntil)
		require.NoError(t, err)

		txadapters.CurrentTimeProvider.CurrentTime = afterFirstUntil

		l, err := txadapters.WantList.List()
		require.NoError(t, err)
		require.NotEmpty(t, l, "if the deadline hasn't passed the value should be returned")

		return nil
	})
	require.NoError(t, err)
}

func TestWantListRepositoryShorterUntilDoesNotOverwriteLongerUntil(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		firstUntil := time.Now()
		afterFirstUntil := firstUntil.Add(fixtures.SomeDuration())
		secondUntil := afterFirstUntil.Add(fixtures.SomeDuration())

		err = txadapters.WantList.AddToWantList(fixtures.SomeRefBlob(), secondUntil)
		require.NoError(t, err)

		err = txadapters.WantList.AddToWantList(fixtures.SomeRefBlob(), firstUntil)
		require.NoError(t, err)

		txadapters.CurrentTimeProvider.CurrentTime = afterFirstUntil

		l, err := txadapters.WantList.List()
		require.NoError(t, err)
		require.NotEmpty(t, l, "if the deadline hasn't passed the value should be returned")

		return nil
	})
	require.NoError(t, err)
}

func TestWantListContainsAndDelete(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		until := time.Now()
		now := until.Add(-fixtures.SomeDuration())
		txadapters.CurrentTimeProvider.CurrentTime = now

		id := fixtures.SomeRefBlob()

		ok, err := txadapters.WantList.WantListContains(id)
		require.NoError(t, err)
		require.False(t, ok)

		err = txadapters.WantList.AddToWantList(id, until)
		require.NoError(t, err)

		ok, err = txadapters.WantList.WantListContains(id)
		require.NoError(t, err)
		require.True(t, ok)

		err = txadapters.WantList.DeleteFromWantList(id)
		require.NoError(t, err)

		ok, err = txadapters.WantList.WantListContains(id)
		require.NoError(t, err)
		require.False(t, ok)

		return nil
	})
	require.NoError(t, err)
}
