package replication_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/internal/mocks"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepliesToWantsWithHas(t *testing.T) {
	p := newTestWantsProcess()

	ctx := fixtures.TestContext(t)
	outgoingCh := make(chan messages.BlobWithSizeOrWantDistance)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), mocks.NewConnectionMock(ctx))
	p.WantsProcess.AddOutgoing(ctx, outgoingCh, peer)

	incomingCh := make(chan messages.BlobWithSizeOrWantDistance)
	p.WantsProcess.AddIncoming(ctx, incomingCh)

	blobId := fixtures.SomeRefBlob()
	p.BlobStorage.MockBlob(blobId, fixtures.SomeBytes())

	go func() {
		select {
		case outgoingCh <- messages.MustNewBlobWithWantDistance(blobId, blobs.NewWantDistanceLocal()):
			return
		case <-ctx.Done():
			panic("context done")
		}
	}()

	select {
	case sent := <-incomingCh:
		require.Equal(t, blobId, sent.Id())
	case <-time.After(1 * time.Second):
		t.Fatal("timeout")
	}
}

func TestRepliesToWantsWithHasIfIncomingChannelConnectedLater(t *testing.T) {
	p := newTestWantsProcess()

	ctx := fixtures.TestContext(t)
	outgoingCh := make(chan messages.BlobWithSizeOrWantDistance)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), mocks.NewConnectionMock(ctx))
	p.WantsProcess.AddOutgoing(ctx, outgoingCh, peer)

	blobId := fixtures.SomeRefBlob()
	p.BlobStorage.MockBlob(blobId, fixtures.SomeBytes())

	select {
	case outgoingCh <- messages.MustNewBlobWithWantDistance(blobId, blobs.NewWantDistanceLocal()):
	case <-ctx.Done():
		panic("context done")
	}

	incomingCh := make(chan messages.BlobWithSizeOrWantDistance)
	p.WantsProcess.AddIncoming(ctx, incomingCh)

	select {
	case sent := <-incomingCh:
		require.Equal(t, blobId, sent.Id())
	case <-time.After(1 * time.Second):
		t.Fatal("timeout")
	}
}

func TestPassesHasToDownloader(t *testing.T) {
	p := newTestWantsProcess()

	ctx, cancel := context.WithTimeout(fixtures.TestContext(t), 5*time.Second)
	defer cancel()

	outgoingCh := make(chan messages.BlobWithSizeOrWantDistance)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), mocks.NewConnectionMock(ctx))
	p.WantsProcess.AddOutgoing(ctx, outgoingCh, peer)

	blobId := fixtures.SomeRefBlob()
	blobSize := fixtures.SomeSize()

	select {
	case outgoingCh <- messages.MustNewBlobWithSize(blobId, blobSize):
	case <-ctx.Done():
		t.Fatal("context done")
	}

	require.Eventually(t,
		func() bool {
			return assert.ObjectsAreEqual(
				[]onHasReceivedCall{
					{
						Peer: peer,
						Blob: blobId,
						Size: blobSize,
					},
				},
				p.HasHandler.OnHasReceivedCalls(),
			)
		},
		1*time.Second,
		10*time.Millisecond,
	)
}

func TestSecondsLocalWants(t *testing.T) {
	p := newTestWantsProcess()

	ctx, cancel := context.WithTimeout(fixtures.TestContext(t), 5*time.Second)
	defer cancel()

	p.WantedBlobsProvider.WantedBlobs = []refs.Blob{
		fixtures.SomeRefBlob(),
	}

	ch := make(chan messages.BlobWithSizeOrWantDistance)

	p.WantsProcess.AddIncoming(ctx, ch)

	var localWants []messages.BlobWithSizeOrWantDistance

drainchannel:
	for {
		select {
		case wants := <-ch:
			localWants = append(localWants, wants)
		case <-time.After(100 * time.Millisecond):
			break drainchannel
		}
	}

	require.Equal(t, len(p.WantedBlobsProvider.WantedBlobs), len(localWants))
}

func TestBuildWantList(t *testing.T) {
	blob1 := fixtures.SomeRefBlob()
	blob2 := fixtures.SomeRefBlob()
	blob3 := fixtures.SomeRefBlob()
	blob4 := fixtures.SomeRefBlob()

	testCases := []struct {
		Name        string
		WantedBlobs []refs.Blob
		BlobsToPush []refs.Blob
		Result      []blobs.WantedBlob
	}{
		{
			Name: "wanted_blobs",
			WantedBlobs: []refs.Blob{
				blob1,
				blob2,
			},
			BlobsToPush: nil,
			Result: []blobs.WantedBlob{
				{
					Id:       blob1,
					Distance: blobs.NewWantDistanceLocal(),
				},
				{
					Id:       blob2,
					Distance: blobs.NewWantDistanceLocal(),
				},
			},
		},
		{
			Name:        "blobs_to_push",
			WantedBlobs: nil,
			BlobsToPush: []refs.Blob{
				blob1,
				blob2,
			},
			Result: []blobs.WantedBlob{
				{
					Id:       blob1,
					Distance: blobs.NewWantDistanceLocal(),
				},
				{
					Id:       blob2,
					Distance: blobs.NewWantDistanceLocal(),
				},
			},
		},
		{
			Name: "wanted_blobs_and_blobs_to_push",
			WantedBlobs: []refs.Blob{
				blob1,
				blob2,
			},
			BlobsToPush: []refs.Blob{
				blob3,
				blob4,
			},
			Result: []blobs.WantedBlob{
				{
					Id:       blob1,
					Distance: blobs.NewWantDistanceLocal(),
				},
				{
					Id:       blob2,
					Distance: blobs.NewWantDistanceLocal(),
				},
				{
					Id:       blob3,
					Distance: blobs.NewWantDistanceLocal(),
				},
				{
					Id:       blob4,
					Distance: blobs.NewWantDistanceLocal(),
				},
			},
		},
		{
			Name: "duplicates",
			WantedBlobs: []refs.Blob{
				blob1,
				blob2,
			},
			BlobsToPush: []refs.Blob{
				blob2,
				blob3,
			},
			Result: []blobs.WantedBlob{
				{
					Id:       blob1,
					Distance: blobs.NewWantDistanceLocal(),
				},
				{
					Id:       blob2,
					Distance: blobs.NewWantDistanceLocal(),
				},
				{
					Id:       blob3,
					Distance: blobs.NewWantDistanceLocal(),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			wantList, err := replication.BuildWantList(testCase.WantedBlobs, testCase.BlobsToPush)
			require.NoError(t, err)

			got := wantList.List()
			expected := testCase.Result

			compare := func(a, b blobs.WantedBlob) bool {
				return a.Id.String() < b.Id.String()
			}

			internal.SortSlice(got, compare)
			internal.SortSlice(expected, compare)

			require.Equal(t, expected, got)
		})
	}
}

type TestWantsProcess struct {
	WantsProcess        *replication.WantsProcess
	WantedBlobsProvider *wantedBlobsProviderMock
	BlobStorage         *mocks.BlobStorageMock
	HasHandler          *hasHandlerMock
}

func newTestWantsProcess() TestWantsProcess {
	wantedBlobsProvider := newWantedBlobsProviderMock()
	blobsThatShouldBePushedProvider := newBlobsThatShouldBePushedProviderMock()
	hasHandler := newHasHandlerMock()
	blobStorage := mocks.NewBlobStorageMock()
	logger := logging.NewDevNullLogger()

	process := replication.NewWantsProcess(
		wantedBlobsProvider,
		blobsThatShouldBePushedProvider,
		blobStorage,
		hasHandler,
		logger,
	)

	return TestWantsProcess{
		WantsProcess:        process,
		WantedBlobsProvider: wantedBlobsProvider,
		BlobStorage:         blobStorage,
		HasHandler:          hasHandler,
	}
}

type wantedBlobsProviderMock struct {
	WantedBlobs []refs.Blob
}

func newWantedBlobsProviderMock() *wantedBlobsProviderMock {
	return &wantedBlobsProviderMock{}
}

func (w wantedBlobsProviderMock) GetWantedBlobs() ([]refs.Blob, error) {
	return w.WantedBlobs, nil
}

type blobsThatShouldBePushedProviderMock struct {
	BlobsThatShouldBePushed []refs.Blob
}

func newBlobsThatShouldBePushedProviderMock() *blobsThatShouldBePushedProviderMock {
	return &blobsThatShouldBePushedProviderMock{}
}

func (b blobsThatShouldBePushedProviderMock) GetBlobsThatShouldBePushed() ([]refs.Blob, error) {
	return b.BlobsThatShouldBePushed, nil
}

type hasHandlerMock struct {
	onHasReceivedCalls []onHasReceivedCall
	lock               sync.Mutex
}

func newHasHandlerMock() *hasHandlerMock {
	return &hasHandlerMock{}
}

func (d *hasHandlerMock) OnHasReceived(ctx context.Context, peer transport.Peer, blob refs.Blob, size blobs.Size) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.onHasReceivedCalls = append(d.onHasReceivedCalls, onHasReceivedCall{
		Peer: peer,
		Blob: blob,
		Size: size,
	})
}

func (d *hasHandlerMock) OnHasReceivedCalls() []onHasReceivedCall {
	d.lock.Lock()
	defer d.lock.Unlock()

	tmp := make([]onHasReceivedCall, len(d.onHasReceivedCalls))
	copy(tmp, d.onHasReceivedCalls)
	return tmp
}

type onHasReceivedCall struct {
	Peer transport.Peer
	Blob refs.Blob
	Size blobs.Size
}
