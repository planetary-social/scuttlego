package graph_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestContacts(t *testing.T) {
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

	require.Equal(t,
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
		"social graph should have returned results sorted by distance and filter out nodes that are too far away",
	)

	require.True(t, g.HasContact(local))
	require.True(t, g.HasContact(a))
	require.True(t, g.HasContact(b))
	require.False(t, g.HasContact(c))
}

type StorageMock struct {
	contacts map[string][]*feeds.Contact
}

func (s StorageMock) GetContacts(node refs.Identity) ([]*feeds.Contact, error) {
	return s.contacts[node.String()], nil
}
