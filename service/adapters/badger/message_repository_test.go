package badger_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/stretchr/testify/require"
)

func TestMessageRepository_CountEmpty(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		n, err := adapters.MessageRepository.Count()
		require.NoError(t, err)
		require.Equal(t, 0, n)

		return nil
	})
	require.NoError(t, err)
}

func TestMessageRepository_GetNoMessage(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		_, err := adapters.MessageRepository.Get(fixtures.SomeRefMessage())
		require.EqualError(t, err, "message not found: Key not found")

		return nil
	})
	require.NoError(t, err)
}

func TestMessageRepository_Put_Get(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.MessageRepository.Put(msg)
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
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

func TestMessageRepository_Delete(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.MessageRepository.Put(msg)
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		_, err = adapters.MessageRepository.Get(msg.Id())
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.MessageRepository.Delete(msg.Id())
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		_, err = adapters.MessageRepository.Get(msg.Id())
		require.EqualError(t, err, "message not found: Key not found")

		return nil
	})
	require.NoError(t, err)
}
