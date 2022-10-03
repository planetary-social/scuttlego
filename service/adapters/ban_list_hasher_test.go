package adapters

import (
	"bytes"
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/stretchr/testify/require"
)

func TestBanListHasher_HashForFeed_ReturnsAnInitializedHash(t *testing.T) {
	hasher := NewBanListHasher()

	hash, err := hasher.HashForFeed(fixtures.SomeRefFeed())
	require.NoError(t, err)
	require.False(t, hash.IsZero())
}

func TestBanListHasher_HashForFeed_ReturnsDifferentHashesForDifferentFeedRefs(t *testing.T) {
	hasher := NewBanListHasher()

	a, err := hasher.HashForFeed(fixtures.SomeRefFeed())
	require.NoError(t, err)

	b, err := hasher.HashForFeed(fixtures.SomeRefFeed())
	require.NoError(t, err)

	require.False(t, bytes.Equal(a.Bytes(), b.Bytes()))
}
