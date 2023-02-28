package graph_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/domain/bans"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/stretchr/testify/require"
)

func TestCachedBanList(t *testing.T) {
	hasher := mocks.NewBanListHasherMock()
	lister := newBanListListerMock()

	feed1 := fixtures.SomeRefFeed()
	hash1 := fixtures.SomeBanListHash()

	feed2 := fixtures.SomeRefFeed()
	hash2 := fixtures.SomeBanListHash()

	feed3 := fixtures.SomeRefFeed()
	hash3 := fixtures.SomeBanListHash()

	hasher.Mock(feed1, hash1)
	hasher.Mock(feed2, hash2)
	hasher.Mock(feed3, hash3)

	lister.ListReturnValue = []bans.Hash{
		hash1,
		hash2,
	}

	list, err := graph.NewCachedBanList(hasher, lister)
	require.NoError(t, err)

	ok, err := list.ContainsFeed(feed1)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = list.ContainsFeed(feed2)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = list.ContainsFeed(feed3)
	require.NoError(t, err)
	require.False(t, ok)
}

type banListListerMock struct {
	ListReturnValue []bans.Hash
}

func newBanListListerMock() *banListListerMock {
	return &banListListerMock{}
}

func (b banListListerMock) List() ([]bans.Hash, error) {
	return b.ListReturnValue, nil
}
