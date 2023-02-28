package badger_test

import (
	"sort"
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
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
	blob1 := feeds.MustNewBlobToSave(
		fixtures.SomeRefBlob(),
	)

	msgRef2 := fixtures.SomeRefMessage()
	blob2 := feeds.MustNewBlobToSave(
		fixtures.SomeRefBlob(),
	)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.BlobRepository.Put(msgRef1, blob1)
		require.NoError(t, err)

		err = adapters.BlobRepository.Put(msgRef2, blob2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		msg1Blobs, err := adapters.BlobRepository.ListBlobs(msgRef1)
		require.NoError(t, err)
		require.Equal(t,
			[]refs.Blob{
				blob1.Ref(),
			},
			msg1Blobs,
		)

		blobMsgs, err := adapters.BlobRepository.ListMessages(blob1.Ref())
		require.NoError(t, err)
		require.Equal(t,
			[]refs.Message{msgRef1},
			blobMsgs,
		)

		msg2Blobs, err := adapters.BlobRepository.ListBlobs(msgRef2)
		require.NoError(t, err)
		require.Equal(t,
			[]refs.Blob{
				blob2.Ref(),
			},
			msg2Blobs,
		)

		blobMsgs, err = adapters.BlobRepository.ListMessages(blob2.Ref())
		require.NoError(t, err)
		require.Equal(t,
			[]refs.Message{msgRef2},
			blobMsgs,
		)

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

		blobMsgs, err := adapters.BlobRepository.ListMessages(blob1.Ref())
		require.NoError(t, err)
		require.Empty(t, blobMsgs)

		msg2Blobs, err := adapters.BlobRepository.ListBlobs(msgRef2)
		require.NoError(t, err)
		require.Equal(t,
			[]refs.Blob{
				blob2.Ref(),
			},
			msg2Blobs,
		)

		blobMsgs, err = adapters.BlobRepository.ListMessages(blob2.Ref())
		require.NoError(t, err)
		require.Equal(t,
			[]refs.Message{msgRef2},
			blobMsgs,
		)

		return nil
	})
	require.NoError(t, err)
}

func TestBlobRepository_DeleteRemovesDataWithoutTouchingOtherEntriesIfMultipleMessagesHaveOneBlob(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	msgRef1 := fixtures.SomeRefMessage()
	msgRef2 := fixtures.SomeRefMessage()

	blob := feeds.MustNewBlobToSave(
		fixtures.SomeRefBlob(),
	)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.BlobRepository.Put(msgRef1, blob)
		require.NoError(t, err)

		err = adapters.BlobRepository.Put(msgRef2, blob)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		msg1Blobs, err := adapters.BlobRepository.ListBlobs(msgRef1)
		require.NoError(t, err)
		require.Equal(t,
			[]refs.Blob{
				blob.Ref(),
			},
			msg1Blobs,
		)

		msg2Blobs, err := adapters.BlobRepository.ListBlobs(msgRef2)
		require.NoError(t, err)
		require.Equal(t,
			[]refs.Blob{
				blob.Ref(),
			},
			msg2Blobs,
		)

		blobMsgs, err := adapters.BlobRepository.ListMessages(blob.Ref())
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

		msg2Blobs, err := adapters.BlobRepository.ListBlobs(msgRef2)
		require.NoError(t, err)
		require.Equal(t,
			[]refs.Blob{
				blob.Ref(),
			},
			msg2Blobs,
		)

		blobMsgs, err := adapters.BlobRepository.ListMessages(blob.Ref())
		require.NoError(t, err)

		expectedBlobMsgs := []refs.Message{msgRef2}

		require.Equal(t,
			expectedBlobMsgs,
			blobMsgs,
		)

		return nil
	})
	require.NoError(t, err)
}
