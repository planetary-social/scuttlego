package bolt_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/bolt"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestFeedRepository_GetFeed_ReturnsAppropriateErrorWhenEmpty(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		_, err = adapters.FeedRepository.GetFeed(fixtures.SomeRefFeed())
		require.ErrorIs(t, err, bolt.ErrFeedNotFound)

		return nil
	})
	require.NoError(t, err)
}
