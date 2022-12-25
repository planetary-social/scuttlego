package notx_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/stretchr/testify/require"
)

func TestNoTxMessageRepository_Count(t *testing.T) {
	ts := di.BuildBadgerNoTxTestAdapters(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	count, err := ts.NoTxTestAdapters.NoTxMessageRepository.Count()
	require.NoError(t, err)
	require.Equal(t, 0, count)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.MessageRepository.Put(msg)
	})
	require.NoError(t, err)

	count, err = ts.NoTxTestAdapters.NoTxMessageRepository.Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
