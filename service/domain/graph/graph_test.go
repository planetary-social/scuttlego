package graph_test

import (
	"sort"
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestContacts_FolloweesAreAddedToTheGraphUpToCertainDepth(t *testing.T) {
	local := fixtures.SomeRefIdentity()

	a := fixtures.SomeRefIdentity()
	b := fixtures.SomeRefIdentity()
	c := fixtures.SomeRefIdentity()

	s := StorageMock{
		contacts: map[string][]*feeds.Contact{
			local.String(): {
				feeds.MustNewContactFromHistory(a, true, false),
			},
			a.String(): {
				feeds.MustNewContactFromHistory(b, true, false),
			},
			b.String(): {
				feeds.MustNewContactFromHistory(c, true, false),
			},
		},
	}

	g, err := graph.NewSocialGraph(local, graph.MustNewHops(2), s)
	require.NoError(t, err)

	requireEqualContacts(t,
		[]graph.Contact{
			{
				local,
				graph.MustNewHops(0),
			},
			{
				a,
				graph.MustNewHops(1),
			},
			{
				b,
				graph.MustNewHops(2),
			},
		},
		g.Contacts(),
	)

	require.True(t, g.HasContact(local))
	require.True(t, g.HasContact(a))
	require.True(t, g.HasContact(b))
	require.False(t, g.HasContact(c))
}

func TestContacts_SmallerNumberOfHopsTakesPriority(t *testing.T) {
	local := fixtures.SomeRefIdentity()

	a := fixtures.SomeRefIdentity()
	b := fixtures.SomeRefIdentity()

	s := StorageMock{
		contacts: map[string][]*feeds.Contact{
			local.String(): {
				feeds.MustNewContactFromHistory(a, true, false),
				feeds.MustNewContactFromHistory(b, true, false),
			},
			a.String(): {
				feeds.MustNewContactFromHistory(b, true, false),
			},
		},
	}

	g, err := graph.NewSocialGraph(local, graph.MustNewHops(2), s)
	require.NoError(t, err)

	requireEqualContacts(t,
		[]graph.Contact{
			{
				local,
				graph.MustNewHops(0),
			},
			{
				a,
				graph.MustNewHops(1),
			},
			{
				b,
				graph.MustNewHops(1),
			},
		},
		g.Contacts(),
	)
}

func TestContacts_EdgesWhichAreBlockedOrNotSetAsFollowingAreNotFollowed(t *testing.T) {
	local := fixtures.SomeRefIdentity()

	a := fixtures.SomeRefIdentity()
	b := fixtures.SomeRefIdentity()
	c := fixtures.SomeRefIdentity()
	d := fixtures.SomeRefIdentity()
	e := fixtures.SomeRefIdentity()

	s := StorageMock{
		contacts: map[string][]*feeds.Contact{
			local.String(): {
				feeds.MustNewContactFromHistory(a, true, false),
			},
			a.String(): {
				feeds.MustNewContactFromHistory(b, true, false),
				feeds.MustNewContactFromHistory(c, false, true),
				feeds.MustNewContactFromHistory(d, false, false),
				feeds.MustNewContactFromHistory(e, true, true),
			},
		},
	}

	g, err := graph.NewSocialGraph(local, graph.MustNewHops(2), s)
	require.NoError(t, err)

	requireEqualContacts(t,
		[]graph.Contact{
			{
				local,
				graph.MustNewHops(0),
			},
			{
				a,
				graph.MustNewHops(1),
			},
			{
				b,
				graph.MustNewHops(2),
			},
		},
		g.Contacts(),
	)
}

func TestContacts_LocalBlockingTakesPriorityAndAlwaysExcludesFeeds(t *testing.T) {
	local := fixtures.SomeRefIdentity()

	a := fixtures.SomeRefIdentity()
	b := fixtures.SomeRefIdentity()
	c := fixtures.SomeRefIdentity()
	d := fixtures.SomeRefIdentity()
	e := fixtures.SomeRefIdentity()

	s := StorageMock{
		contacts: map[string][]*feeds.Contact{
			local.String(): {
				feeds.MustNewContactFromHistory(a, true, false),
				feeds.MustNewContactFromHistory(b, true, false),
				feeds.MustNewContactFromHistory(c, false, true),
				feeds.MustNewContactFromHistory(d, false, false),
				feeds.MustNewContactFromHistory(e, true, true),
			},
			a.String(): {
				feeds.MustNewContactFromHistory(b, true, false),
				feeds.MustNewContactFromHistory(c, true, false),
				feeds.MustNewContactFromHistory(d, true, false),
				feeds.MustNewContactFromHistory(e, true, false),
			},
		},
	}

	g, err := graph.NewSocialGraph(local, graph.MustNewHops(2), s)
	require.NoError(t, err)

	requireEqualContacts(t,
		[]graph.Contact{
			{
				local,
				graph.MustNewHops(0),
			},
			{
				a,
				graph.MustNewHops(1),
			},
			{
				b,
				graph.MustNewHops(1),
			},
			{
				d,
				graph.MustNewHops(2),
			},
		},
		g.Contacts(),
	)
}

func requireEqualContacts(t *testing.T, a []graph.Contact, b []graph.Contact) {
	sort.Slice(a, func(i, j int) bool {
		return a[i].Id.String() < a[j].Id.String()
	})
	sort.Slice(b, func(i, j int) bool {
		return b[i].Id.String() < b[j].Id.String()
	})
	require.Equal(t, a, b)
}

type StorageMock struct {
	contacts map[string][]*feeds.Contact
}

func (s StorageMock) GetContacts(node refs.Identity) ([]*feeds.Contact, error) {
	return s.contacts[node.String()], nil
}
