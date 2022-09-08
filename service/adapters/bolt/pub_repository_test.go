package bolt_test

import (
	"fmt"
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestPubRepository_Delete_SamePubSameAuthorSameAddressSameMessage(t *testing.T) {
	db := fixtures.Bolt(t)

	pub := feeds.NewPubToSave(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefMessage(),
		content.MustNewPub(fixtures.SomeRefIdentity(), fixtures.SomeString(), fixtures.SomeNonNegativeInt()),
	)

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	pubByMessageBucketExists(t, db, pub.Message(), true)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pub.Content().Key().String()),
			utils.BucketName("addresses"),
			utils.BucketName(fmt.Sprintf("%s:%d", pub.Content().Host(), pub.Content().Port())),
			utils.BucketName("sources"),
			utils.BucketName(pub.Who().String()),
			utils.BucketName("messages"),
		},
		true,
	)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		return adapters.PubRepository.Delete(pub.Message())
	})
	require.NoError(t, err)

	pubByMessageBucketExists(t, db, pub.Message(), false)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pub.Content().Key().String()),
		},
		false,
	)
}

func TestPubRepository_Delete_SamePubSameAuthorSameAddressDifferentMessage(t *testing.T) {
	db := fixtures.Bolt(t)

	authorIdenRef := fixtures.SomeRefIdentity()
	pub := content.MustNewPub(fixtures.SomeRefIdentity(), fixtures.SomeString(), fixtures.SomeNonNegativeInt())

	pub1 := feeds.NewPubToSave(
		authorIdenRef,
		fixtures.SomeRefMessage(),
		pub,
	)

	pub2 := feeds.NewPubToSave(
		authorIdenRef,
		fixtures.SomeRefMessage(),
		pub,
	)

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub1)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	pubByMessageBucketExists(t, db, pub1.Message(), true)
	pubByMessageBucketExists(t, db, pub2.Message(), true)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pub.Key().String()),
			utils.BucketName("addresses"),
			utils.BucketName(fmt.Sprintf("%s:%d", pub.Host(), pub.Port())),
			utils.BucketName("sources"),
			utils.BucketName(authorIdenRef.String()),
			utils.BucketName("messages"),
		},
		true,
	)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		return adapters.PubRepository.Delete(pub1.Message())
	})
	require.NoError(t, err)

	pubByMessageBucketExists(t, db, pub1.Message(), false)
	pubByMessageBucketExists(t, db, pub2.Message(), true)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pub.Key().String()),
			utils.BucketName("addresses"),
			utils.BucketName(fmt.Sprintf("%s:%d", pub.Host(), pub.Port())),
			utils.BucketName("sources"),
			utils.BucketName(authorIdenRef.String()),
			utils.BucketName("messages"),
		},
		true,
	)
}

func TestPubRepository_Delete_SamePubDifferentAuthorSameAddressDifferentMessage(t *testing.T) {
	db := fixtures.Bolt(t)

	pub := content.MustNewPub(fixtures.SomeRefIdentity(), fixtures.SomeString(), fixtures.SomeNonNegativeInt())

	pub1 := feeds.NewPubToSave(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefMessage(),
		pub,
	)

	pub2 := feeds.NewPubToSave(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefMessage(),
		pub,
	)

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub1)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	pubByMessageBucketExists(t, db, pub1.Message(), true)
	pubByMessageBucketExists(t, db, pub2.Message(), true)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pub.Key().String()),
			utils.BucketName("addresses"),
			utils.BucketName(fmt.Sprintf("%s:%d", pub.Host(), pub.Port())),
			utils.BucketName("sources"),
			utils.BucketName(pub1.Who().String()),
		},
		true,
	)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pub.Key().String()),
			utils.BucketName("addresses"),
			utils.BucketName(fmt.Sprintf("%s:%d", pub.Host(), pub.Port())),
			utils.BucketName("sources"),
			utils.BucketName(pub2.Who().String()),
		},
		true,
	)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		return adapters.PubRepository.Delete(pub1.Message())
	})
	require.NoError(t, err)

	pubByMessageBucketExists(t, db, pub1.Message(), false)
	pubByMessageBucketExists(t, db, pub2.Message(), true)

	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pub.Key().String()),
			utils.BucketName("addresses"),
			utils.BucketName(fmt.Sprintf("%s:%d", pub.Host(), pub.Port())),
			utils.BucketName("sources"),
			utils.BucketName(pub1.Who().String()),
		},
		false,
	)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pub.Key().String()),
			utils.BucketName("addresses"),
			utils.BucketName(fmt.Sprintf("%s:%d", pub.Host(), pub.Port())),
			utils.BucketName("sources"),
			utils.BucketName(pub2.Who().String()),
		},
		true,
	)
}

func TestPubRepository_Delete_SamePubDifferentAuthorDifferentAddressDifferentMessage(t *testing.T) {
	db := fixtures.Bolt(t)

	pubIden := fixtures.SomeRefIdentity()

	pub1 := feeds.NewPubToSave(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefMessage(),
		content.MustNewPub(pubIden, fixtures.SomeString(), fixtures.SomeNonNegativeInt()),
	)

	pub2 := feeds.NewPubToSave(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefMessage(),
		content.MustNewPub(pubIden, fixtures.SomeString(), fixtures.SomeNonNegativeInt()),
	)

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub1)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	pubByMessageBucketExists(t, db, pub1.Message(), true)
	pubByMessageBucketExists(t, db, pub2.Message(), true)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pubIden.String()),
			utils.BucketName("addresses"),
			utils.BucketName(fmt.Sprintf("%s:%d", pub1.Content().Host(), pub1.Content().Port())),
		},
		true,
	)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pubIden.String()),
			utils.BucketName("addresses"),
			utils.BucketName(fmt.Sprintf("%s:%d", pub2.Content().Host(), pub2.Content().Port())),
		},
		true,
	)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		return adapters.PubRepository.Delete(pub1.Message())
	})
	require.NoError(t, err)

	pubByMessageBucketExists(t, db, pub1.Message(), false)
	pubByMessageBucketExists(t, db, pub2.Message(), true)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pubIden.String()),
			utils.BucketName("addresses"),
			utils.BucketName(fmt.Sprintf("%s:%d", pub1.Content().Host(), pub1.Content().Port())),
		},
		false,
	)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pubIden.String()),
			utils.BucketName("addresses"),
			utils.BucketName(fmt.Sprintf("%s:%d", pub2.Content().Host(), pub2.Content().Port())),
		},
		true,
	)
}

func TestPubRepository_Delete_DifferentPubDifferentAuthorDifferentAddressDifferentMessage(t *testing.T) {
	db := fixtures.Bolt(t)

	pub1 := feeds.NewPubToSave(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefMessage(),
		content.MustNewPub(fixtures.SomeRefIdentity(), fixtures.SomeString(), fixtures.SomeNonNegativeInt()),
	)

	pub2 := feeds.NewPubToSave(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefMessage(),
		content.MustNewPub(fixtures.SomeRefIdentity(), fixtures.SomeString(), fixtures.SomeNonNegativeInt()),
	)

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub1)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	pubByMessageBucketExists(t, db, pub1.Message(), true)
	pubByMessageBucketExists(t, db, pub2.Message(), true)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pub1.Content().Key().String()),
		},
		true,
	)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pub2.Content().Key().String()),
		},
		true,
	)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		return adapters.PubRepository.Delete(pub1.Message())
	})
	require.NoError(t, err)

	pubByMessageBucketExists(t, db, pub1.Message(), false)
	pubByMessageBucketExists(t, db, pub2.Message(), true)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pub1.Content().Key().String()),
		},
		false,
	)
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_pub"),
			utils.BucketName(pub2.Content().Key().String()),
		},
		true,
	)
}

func pubByMessageBucketExists(t *testing.T, db *bbolt.DB, msgRef refs.Message, exists bool) {
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("pubs"),
			utils.BucketName("by_message"),
			utils.BucketName(msgRef.String()),
		},
		exists,
	)
}
