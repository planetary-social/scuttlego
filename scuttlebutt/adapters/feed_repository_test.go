package adapters_test

// todo build tags for integration tests

import (
	"github.com/planetary-social/go-ssb/cmd/di"
	"go.etcd.io/bbolt"
	"testing"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/scuttlebutt/replication"
	"github.com/stretchr/testify/require"
)

func TestGetFeed_returns_appropriate_error_when_empty(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildAdaptersForTest(tx)
		require.NoError(t, err)

		_, err = adapters.Feed.GetFeed(fixtures.SomeRefFeed())
		require.ErrorIs(t, err, replication.ErrFeedNotFound)

		return nil
	})
	require.NoError(t, err)
}
