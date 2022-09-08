package bolt_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestBlobRepository_DeleteShouldRemoveTheByBlobBucketIfNoMoreMessagesReferencesTheBlob(t *testing.T) {
	db := fixtures.Bolt(t)

	msgRef1 := fixtures.SomeRefMessage()
	blobs1 := feeds.NewBlobToSave(
		[]refs.Blob{
			fixtures.SomeRefBlob(),
		},
	)

	msgRef2 := fixtures.SomeRefMessage()
	blobs2 := feeds.NewBlobToSave(
		[]refs.Blob{
			fixtures.SomeRefBlob(),
		},
	)

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		err = adapters.BlobRepository.Put(msgRef1, blobs1)
		require.NoError(t, err)

		err = adapters.BlobRepository.Put(msgRef2, blobs2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	byMessageBucketExists(t, db, msgRef1, true)
	byBlobBucketsExist(t, db, blobs1, true)

	byMessageBucketExists(t, db, msgRef2, true)
	byBlobBucketsExist(t, db, blobs2, true)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		return adapters.BlobRepository.Delete(msgRef1)
	})
	require.NoError(t, err)

	byMessageBucketExists(t, db, msgRef1, false)
	byBlobBucketsExist(t, db, blobs1, false)

	byMessageBucketExists(t, db, msgRef2, true)
	byBlobBucketsExist(t, db, blobs2, true)
}

func TestBlobRepository_DeleteShouldNotRemoveTheByBlobBucketIfAnotherMessageStillReferencesTheBlob(t *testing.T) {
	db := fixtures.Bolt(t)

	msgRef1 := fixtures.SomeRefMessage()
	msgRef2 := fixtures.SomeRefMessage()

	blobs := feeds.NewBlobToSave(
		[]refs.Blob{
			fixtures.SomeRefBlob(),
		},
	)

	err := db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		err = adapters.BlobRepository.Put(msgRef1, blobs)
		require.NoError(t, err)

		err = adapters.BlobRepository.Put(msgRef2, blobs)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	byMessageBucketExists(t, db, msgRef1, true)
	byMessageBucketExists(t, db, msgRef2, true)

	byBlobBucketsExist(t, db, blobs, true)

	err = db.Update(func(tx *bbolt.Tx) error {
		adapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		return adapters.BlobRepository.Delete(msgRef1)
	})
	require.NoError(t, err)

	byMessageBucketExists(t, db, msgRef1, false)
	byMessageBucketExists(t, db, msgRef2, true)

	byBlobBucketsExist(t, db, blobs, true)
}

func byMessageBucketExists(t *testing.T, db *bbolt.DB, msgRef refs.Message, exists bool) {
	requireBucketExistsNoTx(
		t,
		db,
		[]utils.BucketName{
			utils.BucketName("blobs"),
			utils.BucketName("by_message"),
			utils.BucketName(msgRef.String()),
		},
		exists,
	)
}

func byBlobBucketsExist(t *testing.T, db *bbolt.DB, blobs feeds.BlobToSave, exists bool) {
	err := db.View(func(tx *bbolt.Tx) error {
		for _, b := range blobs.Blobs() {
			requireBucketExists(
				t,
				tx,
				[]utils.BucketName{
					utils.BucketName("blobs"),
					utils.BucketName("by_blob"),
					utils.BucketName(b.String()),
				},
				exists,
			)
		}

		return nil
	})
	require.NoError(t, err)
}

func requireBucketExistsNoTx(t *testing.T, db *bbolt.DB, bucket []utils.BucketName, exists bool) {
	err := db.View(func(tx *bbolt.Tx) error {
		requireBucketExists(
			t,
			tx,
			bucket,
			exists,
		)
		return nil
	})
	require.NoError(t, err)
}

func requireBucketExists(t *testing.T, tx *bbolt.Tx, bucket []utils.BucketName, exists bool) {
	b, err := utils.GetBucket(tx, bucket)
	require.NoError(t, err)
	require.Equal(t, exists, b != nil)
}
