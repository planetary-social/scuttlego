package bolt_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

// todo more tests (eg. blocks should prevent get contects from returning values)

func TestSocialGraphRepository_RemoveDropsContactDataForTheSpecifiedFeed(t *testing.T) {
	db := fixtures.Bolt(t)

	iden1 := fixtures.SomeRefIdentity()
	iden2 := fixtures.SomeRefIdentity()

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		err = txadapters.SocialGraphRepository.Follow(iden1, fixtures.SomeRefIdentity())
		require.NoError(t, err)

		err = txadapters.SocialGraphRepository.Follow(iden2, fixtures.SomeRefIdentity())
		require.NoError(t, err)

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
