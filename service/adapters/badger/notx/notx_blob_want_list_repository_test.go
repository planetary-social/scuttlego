package notx_test

import (
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/stretchr/testify/require"
)

func TestNoTxBlobWantListRepository_ListReturnsEmptyWantListIfDatabaseIsEmpty(t *testing.T) {
	ts := di.BuildBadgerNoTxTestAdapters(t)

	wantlist, err := ts.NoTxTestAdapters.NoTxBlobWantListRepository.List()
	require.NoError(t, err)
	require.Empty(t, wantlist.List())
}

func TestNoTxBlobWantListRepository_ListReturnsResultsNonEmptyWantListIfDatabaseIsNotEmpty(t *testing.T) {
	ts := di.BuildBadgerNoTxTestAdapters(t)

	fixtures.SomeRefBlob()

	now := time.Now()
	blobRef := fixtures.SomeRefBlob()

	ts.Dependencies.CurrentTimeProvider.CurrentTime = now

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.BlobWantListRepository.Add(blobRef, now.Add(1*time.Second))
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	wantlist, err := ts.NoTxTestAdapters.NoTxBlobWantListRepository.List()
	require.NoError(t, err)
	require.Equal(t,
		[]blobs.WantedBlob{
			{
				Id:       blobRef,
				Distance: blobs.MustNewWantDistance(1),
			},
		},
		wantlist.List(),
	)
}
