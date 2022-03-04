package adapters_test

// todo build tags for integration tests

import (
	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/scuttlebutt/adapters"
	"github.com/planetary-social/go-ssb/scuttlebutt/replication"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetFeed_returns_appropriate_error_when_empty(t *testing.T) {
	db := fixtures.Bolt(t)

	s := adapters.NewBoltFeedStorage(db, nil) // todo wire

	_, err := s.GetFeed(fixtures.SomeRefFeed())
	require.ErrorIs(t, err, replication.ErrFeedNotFound)
}
