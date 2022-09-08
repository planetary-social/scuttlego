package bolt_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestMessageRepository_CountEmpty(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.View(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		n, err := adapters.MessageRepository.Count()
		require.NoError(t, err)
		require.Equal(t, 0, n)

		return nil
	})
	require.NoError(t, err)
}

func TestMessageRepository_GetNoMessage(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.View(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		_, err = adapters.MessageRepository.Get(fixtures.SomeRefMessage())
		require.EqualError(t, err, "message not found")

		return nil
	})
	require.NoError(t, err)
}

func TestMessageRepository_Put_Get(t *testing.T) {
	db := fixtures.Bolt(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		return adapters.MessageRepository.Put(msg)
	})
	require.NoError(t, err)

	err = db.View(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		retrievedMessage, err := adapters.MessageRepository.Get(msg.Id())
		require.NoError(t, err)
		require.Equal(t, retrievedMessage.Raw(), msg.Raw())

		n, err := adapters.MessageRepository.Count()
		require.NoError(t, err)
		require.Equal(t, 1, n)

		return nil
	})
	require.NoError(t, err)
}

func TestReadMessageRepository_Count(t *testing.T) {
	db := fixtures.Bolt(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	a, err := di.BuildTestAdapters(db)
	require.NoError(t, err)

	count, err := a.MessageRepository.Count()
	require.NoError(t, err)
	require.Equal(t, 0, count)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		return adapters.MessageRepository.Put(msg)
	})
	require.NoError(t, err)

	count, err = a.MessageRepository.Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestMessageRepository_Delete(t *testing.T) {
	db := fixtures.Bolt(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		return adapters.MessageRepository.Put(msg)
	})
	require.NoError(t, err)

	err = db.View(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		_, err = adapters.MessageRepository.Get(msg.Id())
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		return adapters.MessageRepository.Delete(msg.Id())
	})
	require.NoError(t, err)

	err = db.View(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		_, err = adapters.MessageRepository.Get(msg.Id())
		require.EqualError(t, err, "message not found")

		return nil
	})
	require.NoError(t, err)
}
