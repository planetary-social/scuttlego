package badger_test

import (
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"sort"
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestBlobRepository_ListingDoesNotReturnErrorsIfBlobOrMessageIsNotKnown(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		blobs, err := adapters.BlobRepository.ListBlobs(fixtures.SomeRefMessage())
		require.NoError(t, err)
		require.Empty(t, blobs)

		msgs, err := adapters.BlobRepository.ListMessages(fixtures.SomeRefBlob())
		require.NoError(t, err)
		require.Empty(t, msgs)

		return nil
	})
	require.NoError(t, err)
}

func TestBlobRepository_DeleteRemovesDataWithoutTouchingOtherEntries(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

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

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.BlobRepository.Put(msgRef1, blobs1)
		require.NoError(t, err)

		err = adapters.BlobRepository.Put(msgRef2, blobs2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		msg1Blobs, err := adapters.BlobRepository.ListBlobs(msgRef1)
		require.NoError(t, err)
		require.Equal(t,
			blobs1.Blobs(),
			msg1Blobs,
		)

		for _, blob := range blobs1.Blobs() {
			blobMsgs, err := adapters.BlobRepository.ListMessages(blob)
			require.NoError(t, err)
			require.Equal(t,
				[]refs.Message{msgRef1},
				blobMsgs,
			)
		}

		msg2Blobs, err := adapters.BlobRepository.ListBlobs(msgRef2)
		require.NoError(t, err)
		require.Equal(t,
			blobs2.Blobs(),
			msg2Blobs,
		)

		for _, blob := range blobs2.Blobs() {
			blobMsgs, err := adapters.BlobRepository.ListMessages(blob)
			require.NoError(t, err)
			require.Equal(t,
				[]refs.Message{msgRef2},
				blobMsgs,
			)
		}

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.BlobRepository.Delete(msgRef1)
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		msg1Blobs, err := adapters.BlobRepository.ListBlobs(msgRef1)
		require.NoError(t, err)
		require.Empty(t, msg1Blobs)

		for _, blob := range blobs1.Blobs() {
			blobMsgs, err := adapters.BlobRepository.ListMessages(blob)
			require.NoError(t, err)
			require.Empty(t, blobMsgs)
		}

		msg2Blobs, err := adapters.BlobRepository.ListBlobs(msgRef2)
		require.NoError(t, err)
		require.Equal(t,
			blobs2.Blobs(),
			msg2Blobs,
		)

		for _, blob := range blobs2.Blobs() {
			blobMsgs, err := adapters.BlobRepository.ListMessages(blob)
			require.NoError(t, err)
			require.Equal(t,
				[]refs.Message{msgRef2},
				blobMsgs,
			)
		}

		return nil
	})
	require.NoError(t, err)
}

func TestBlobRepository_DeleteRemovesDataWithoutTouchingOtherEntriesIfMultipleMessagesHaveOneBlob(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	msgRef1 := fixtures.SomeRefMessage()
	msgRef2 := fixtures.SomeRefMessage()

	blobs := feeds.NewBlobToSave(
		[]refs.Blob{
			fixtures.SomeRefBlob(),
		},
	)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.BlobRepository.Put(msgRef1, blobs)
		require.NoError(t, err)

		err = adapters.BlobRepository.Put(msgRef2, blobs)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	//byMessageBucketExists(t, db, msgRef1, true)
	//byMessageBucketExists(t, db, msgRef2, true)
	//
	//byBlobBucketsExist(t, db, blobs, true)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		msg1Blobs, err := adapters.BlobRepository.ListBlobs(msgRef1)
		require.NoError(t, err)
		require.Equal(t,
			blobs.Blobs(),
			msg1Blobs,
		)

		msg2Blobs, err := adapters.BlobRepository.ListBlobs(msgRef2)
		require.NoError(t, err)
		require.Equal(t,
			blobs.Blobs(),
			msg2Blobs,
		)

		for _, blob := range blobs.Blobs() {
			blobMsgs, err := adapters.BlobRepository.ListMessages(blob)
			require.NoError(t, err)

			expectedBlobMsgs := []refs.Message{msgRef1, msgRef2}

			sort.Slice(blobMsgs, func(i, j int) bool {
				return blobMsgs[i].String() < blobMsgs[j].String()
			})

			sort.Slice(expectedBlobMsgs, func(i, j int) bool {
				return expectedBlobMsgs[i].String() < expectedBlobMsgs[j].String()
			})

			require.Equal(t,
				expectedBlobMsgs,
				blobMsgs,
			)
		}

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.BlobRepository.Delete(msgRef1)
	})
	require.NoError(t, err)

	//byMessageBucketExists(t, db, msgRef1, false)
	//byMessageBucketExists(t, db, msgRef2, true)
	//
	//byBlobBucketsExist(t, db, blobs, true)
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
