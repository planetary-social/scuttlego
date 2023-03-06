package badger_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/di"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestSocialGraphRepository_RemoveDropsContactDataForTheSpecifiedFeed(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	iden1 := fixtures.SomeRefIdentity()
	iden2 := fixtures.SomeRefIdentity()

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		applyContactAction(t, adapters, iden1, fixtures.SomeRefIdentity(), known.ContactActionFollow)
		applyContactAction(t, adapters, iden2, fixtures.SomeRefIdentity(), known.ContactActionFollow)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		contacts, err := adapters.SocialGraphRepository.GetContacts(iden1)
		require.NoError(t, err)
		require.NotEmpty(t, contacts)

		contacts, err = adapters.SocialGraphRepository.GetContacts(iden2)
		require.NoError(t, err)
		require.NotEmpty(t, contacts)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err = adapters.SocialGraphRepository.Remove(iden1)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		contacts, err := adapters.SocialGraphRepository.GetContacts(iden1)
		require.NoError(t, err)
		require.Empty(t, contacts)

		contacts, err = adapters.SocialGraphRepository.GetContacts(iden2)
		require.NoError(t, err)
		require.NotEmpty(t, contacts)

		return nil
	})
	require.NoError(t, err)
}

func TestSocialGraphRepository_GetContacts(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	iden := fixtures.SomeRefIdentity()

	target1 := fixtures.SomeRefIdentity()
	target2 := fixtures.SomeRefIdentity()
	target3 := fixtures.SomeRefIdentity()

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		applyContactAction(t, adapters, iden, target1, known.ContactActionFollow)
		applyContactAction(t, adapters, iden, target2, known.ContactActionFollow)
		applyContactAction(t, adapters, iden, target3, known.ContactActionFollow)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		contacts, err := adapters.SocialGraphRepository.GetContacts(iden)
		require.NoError(t, err)
		sortAndRequireEqualContacts(t,
			[]*feeds.Contact{
				feeds.MustNewContactFromHistory(iden, target1, true, false),
				feeds.MustNewContactFromHistory(iden, target2, true, false),
				feeds.MustNewContactFromHistory(iden, target3, true, false),
			},
			contacts,
		)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		applyContactAction(t, adapters, iden, target1, known.ContactActionBlock)
		applyContactAction(t, adapters, iden, target2, known.ContactActionUnfollow)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		contacts, err := adapters.SocialGraphRepository.GetContacts(iden)
		require.NoError(t, err)
		sortAndRequireEqualContacts(
			t,
			[]*feeds.Contact{
				feeds.MustNewContactFromHistory(iden, target1, true, true),
				feeds.MustNewContactFromHistory(iden, target2, false, false),
				feeds.MustNewContactFromHistory(iden, target3, true, false),
			},
			contacts,
		)

		return nil
	})
	require.NoError(t, err)
}

func BenchmarkSocialGraphRepository_GetContacts(b *testing.B) {
	for _, numberOfFollowees := range []int{0, 1, 10, 100, 1000} {
		b.Run(fmt.Sprintf("followees=%d", numberOfFollowees), func(b *testing.B) {
			ts := di.BuildBadgerTestAdapters(b)
			follower := fixtures.SomeRefIdentity()

			err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
				for i := 0; i < numberOfFollowees; i++ {
					applyContactAction(b, adapters, follower, fixtures.SomeRefIdentity(), known.ContactActionFollow)
				}
				return nil
			})
			require.NoError(b, err)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				err := ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
					contacts, err := adapters.SocialGraphRepository.GetContacts(follower)
					require.NoError(b, err)
					require.Len(b, contacts, numberOfFollowees)

					return nil
				})
				require.NoError(b, err)
			}
		})
	}
}

func sortAndRequireEqualContacts(t *testing.T, a []*feeds.Contact, b []*feeds.Contact) {
	sort.Slice(a, func(i, j int) bool {
		return a[i].Target().String() < a[j].Target().String()
	})

	sort.Slice(b, func(i, j int) bool {
		return b[i].Target().String() < b[j].Target().String()
	})

	require.Equal(t, a, b)
}

func applyContactAction(t testing.TB, adapters badger.TestAdapters, a refs.Identity, b refs.Identity, action known.ContactAction) {
	err := adapters.SocialGraphRepository.UpdateContact(a, b, func(contact *feeds.Contact) error {
		return contact.Update(known.MustNewContactActions([]known.ContactAction{action}))
	})
	require.NoError(t, err)
}
