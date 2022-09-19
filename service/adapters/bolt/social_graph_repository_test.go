package bolt_test

import (
	"sort"
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestSocialGraphRepository_RemoveDropsContactDataForTheSpecifiedFeed(t *testing.T) {
	db := fixtures.Bolt(t)

	iden1 := fixtures.SomeRefIdentity()
	iden2 := fixtures.SomeRefIdentity()

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		applyContactAction(t, txadapters, iden1, fixtures.SomeRefIdentity(), content.ContactActionFollow)
		applyContactAction(t, txadapters, iden2, fixtures.SomeRefIdentity(), content.ContactActionFollow)

		return nil
	})
	require.NoError(t, err)

	err = db.View(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		contacts, err := txadapters.SocialGraphRepository.GetContacts(iden1)
		require.NoError(t, err)
		require.NotEmpty(t, contacts)

		contacts, err = txadapters.SocialGraphRepository.GetContacts(iden2)
		require.NoError(t, err)
		require.NotEmpty(t, contacts)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		err = txadapters.SocialGraphRepository.Remove(iden1)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = db.View(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		contacts, err := txadapters.SocialGraphRepository.GetContacts(iden1)
		require.NoError(t, err)
		require.Empty(t, contacts)

		contacts, err = txadapters.SocialGraphRepository.GetContacts(iden2)
		require.NoError(t, err)
		require.NotEmpty(t, contacts)

		return nil
	})
	require.NoError(t, err)
}

func TestSocialGraphRepository_GetContacts(t *testing.T) {
	db := fixtures.Bolt(t)

	iden := fixtures.SomeRefIdentity()

	target1 := fixtures.SomeRefIdentity()
	target2 := fixtures.SomeRefIdentity()
	target3 := fixtures.SomeRefIdentity()

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		applyContactAction(t, txadapters, iden, target1, content.ContactActionFollow)
		applyContactAction(t, txadapters, iden, target2, content.ContactActionFollow)
		applyContactAction(t, txadapters, iden, target3, content.ContactActionFollow)

		return nil
	})
	require.NoError(t, err)

	err = db.View(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		contacts, err := txadapters.SocialGraphRepository.GetContacts(iden)
		require.NoError(t, err)
		sortAndRequireEqualContacts(t,
			[]*feeds.Contact{
				feeds.MustNewContactFromHistory(target1, true, false),
				feeds.MustNewContactFromHistory(target2, true, false),
				feeds.MustNewContactFromHistory(target3, true, false),
			},
			contacts,
		)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		applyContactAction(t, txadapters, iden, target1, content.ContactActionBlock)
		applyContactAction(t, txadapters, iden, target2, content.ContactActionUnfollow)

		return nil
	})
	require.NoError(t, err)

	err = db.View(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		contacts, err := txadapters.SocialGraphRepository.GetContacts(iden)
		require.NoError(t, err)
		sortAndRequireEqualContacts(
			t,
			[]*feeds.Contact{
				feeds.MustNewContactFromHistory(target1, true, true),
				feeds.MustNewContactFromHistory(target2, false, false),
				feeds.MustNewContactFromHistory(target3, true, false),
			},
			contacts,
		)

		return nil
	})
	require.NoError(t, err)
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

func applyContactAction(t *testing.T, txadapters di.TxTestAdapters, a refs.Identity, b refs.Identity, action content.ContactAction) {
	err := txadapters.SocialGraphRepository.UpdateContact(a, b, func(contact *feeds.Contact) error {
		return contact.Update(content.MustNewContactActions([]content.ContactAction{action}))
	})
	require.NoError(t, err)
}
