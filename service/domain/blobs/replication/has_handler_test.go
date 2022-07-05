package replication_test

import (
	"context"
	"testing"
	"time"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/adapters/mocks"
	"github.com/planetary-social/go-ssb/service/domain/blobs"
	"github.com/planetary-social/go-ssb/service/domain/blobs/replication"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasHandlerTriggersDownloader(t *testing.T) {
	t.Parallel()

	smallSize := blobs.MustNewSize(10)
	largeSize := blobs.MustNewSize(blobs.MaxBlobSize().InBytes() + 10)

	testCases := []struct {
		Name string

		InStorage  bool
		InWantList bool
		Size       blobs.Size

		ShouldTrigger bool
	}{
		{
			Name:          "valid",
			InStorage:     false,
			InWantList:    true,
			Size:          smallSize,
			ShouldTrigger: true,
		},
		{
			Name:          "too_large",
			InStorage:     false,
			InWantList:    true,
			Size:          largeSize,
			ShouldTrigger: false,
		},
		{
			Name:          "already_in_storage",
			InStorage:     true,
			InWantList:    true,
			Size:          smallSize,
			ShouldTrigger: false,
		},
		{
			Name:          "not_in_want_list",
			InStorage:     false,
			InWantList:    false,
			Size:          smallSize,
			ShouldTrigger: false,
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			h := newTestHasHandler()

			ctx := fixtures.TestContext(t)
			peer := transport.NewPeer(fixtures.SomePublicIdentity(), nil)
			blob := fixtures.SomeRefBlob()

			if testCase.InWantList {
				h.WantList.AddBlob(blob)
			}

			if testCase.InStorage {
				h.Storage.MockBlob(blob, fixtures.SomeBytes())
			}

			h.HasHandler.OnHasReceived(ctx, peer, blob, testCase.Size)

			if testCase.ShouldTrigger {
				require.Eventually(t,
					func() bool {
						return assert.ObjectsAreEqual(
							[]downloadCall{
								{
									Peer: peer,
									Blob: blob,
								},
							},
							h.Downloader.DownloadCalls,
						)
					},
					1*time.Second, 10*time.Millisecond)

				require.Eventually(t,
					func() bool {
						return len(h.WantList.List()) == 0
					},
					1*time.Second, 10*time.Millisecond)
			} else {
				<-time.After(1 * time.Second)
				require.Empty(t, h.Downloader.DownloadCalls)
			}
		})
	}
}

func TestHasHandlerRemovesElementFromWantListIfItIsAlreadyInStorage(t *testing.T) {
	t.Parallel()

	h := newTestHasHandler()

	ctx := fixtures.TestContext(t)
	peer := transport.NewPeer(fixtures.SomePublicIdentity(), nil)
	blob := fixtures.SomeRefBlob()

	h.WantList.AddBlob(blob)
	h.Storage.MockBlob(blob, fixtures.SomeBytes())

	h.HasHandler.OnHasReceived(ctx, peer, blob, blobs.MustNewSize(10))

	<-time.After(1 * time.Second)
	require.Empty(t, h.Downloader.DownloadCalls)
	require.Empty(t, h.WantList.List())
}

type testHasHandler struct {
	HasHandler *replication.HasHandler
	WantList   *mocks.WantListRepositoryMock
	Downloader *downloaderMock
	Storage    *mocks.BlobStorageMock
}

func newTestHasHandler() testHasHandler {
	storage := mocks.NewBlobStorageMock()
	wantList := mocks.NewWantListRepositoryMock()
	downloader := newDownloaderMock()

	h := replication.NewHasHandler(
		storage,
		wantList,
		downloader,
		logging.NewDevNullLogger(),
	)

	return testHasHandler{
		HasHandler: h,
		Storage:    storage,
		WantList:   wantList,
		Downloader: downloader,
	}

}

type downloaderMock struct {
	DownloadCalls []downloadCall
}

func newDownloaderMock() *downloaderMock {
	return &downloaderMock{}
}

func (d *downloaderMock) Download(ctx context.Context, peer transport.Peer, blob refs.Blob) error {
	d.DownloadCalls = append(d.DownloadCalls, downloadCall{
		Peer: peer,
		Blob: blob,
	})
	return nil
}

type downloadCall struct {
	Peer transport.Peer
	Blob refs.Blob
}
