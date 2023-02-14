package feeds_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/stretchr/testify/require"
)

func TestContact_AuthorCanBeTheSameAsTarget(t *testing.T) {
	author := fixtures.SomeRefIdentity()

	t.Run("new_contact", func(t *testing.T) {
		_, err := feeds.NewContact(author, author)
		require.NoError(t, err)
	})

	t.Run("new_contact_from_history", func(t *testing.T) {
		_, err := feeds.NewContactFromHistory(author, author, fixtures.SomeBool(), fixtures.SomeBool())
		require.NoError(t, err)
	})
}

func TestNewContactFromHistory(t *testing.T) {
	author := fixtures.SomeRefIdentity()
	target := fixtures.SomeRefIdentity()

	t.Run("following_true", func(t *testing.T) {
		c, err := feeds.NewContactFromHistory(author, target, true, false)
		require.NoError(t, err)

		require.Equal(t, target, c.Target())
		require.True(t, c.Following())
		require.False(t, c.Blocking())
	})

	t.Run("blocking_true", func(t *testing.T) {
		c, err := feeds.NewContactFromHistory(author, target, false, true)
		require.NoError(t, err)

		require.Equal(t, target, c.Target())
		require.False(t, c.Following())
		require.True(t, c.Blocking())
	})
}

func TestContact_Update(t *testing.T) {
	testCases := []struct {
		Name      string
		Actions   []known.ContactAction
		Following bool
		Blocking  bool
	}{
		{
			Name: "follow",
			Actions: []known.ContactAction{
				known.ContactActionFollow,
			},
			Following: true,
			Blocking:  false,
		},
		{
			Name: "unfollow",
			Actions: []known.ContactAction{
				known.ContactActionUnfollow,
			},
			Following: false,
			Blocking:  false,
		},
		{
			Name: "block",
			Actions: []known.ContactAction{
				known.ContactActionBlock,
			},
			Following: false,
			Blocking:  true,
		},
		{
			Name: "unblock",
			Actions: []known.ContactAction{
				known.ContactActionUnblock,
			},
			Following: false,
			Blocking:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			contact, err := feeds.NewContact(
				fixtures.SomeRefIdentity(),
				fixtures.SomeRefIdentity(),
			)
			require.NoError(t, err)

			err = contact.Update(known.MustNewContactActions(testCase.Actions))
			require.NoError(t, err)

			require.Equal(t, testCase.Following, contact.Following())
			require.Equal(t, testCase.Blocking, contact.Blocking())
		})
	}
}

func TestContact_UpdateCorrectlyAppliesAllActions(t *testing.T) {
	actions, err := known.NewContactActions(
		[]known.ContactAction{
			known.ContactActionUnfollow,
			known.ContactActionUnblock,
		},
	)
	require.NoError(t, err)

	contact, err := feeds.NewContactFromHistory(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefIdentity(),
		true,
		true,
	)
	require.NoError(t, err)

	require.True(t, contact.Following())
	require.True(t, contact.Blocking())

	err = contact.Update(actions)
	require.NoError(t, err)

	require.False(t, contact.Following())
	require.False(t, contact.Blocking())
}
