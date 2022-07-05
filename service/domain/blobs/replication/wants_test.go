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
	"github.com/planetary-social/go-ssb/service/domain/messages"
	"github.com/planetary-social/go-ssb/service/domain/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepliesToWantsWithHas(t *testing.T) {
	p := newTestWantsProcess()

	ctx := fixtures.TestContext(t)
	outgoingCh := make(chan messages.BlobWithSizeOrWantDistance)
	peer := transport.NewPeer(fixtures.SomePublicIdentity(), nil)
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
	peer := transport.NewPeer(fixtures.SomePublicIdentity(), nil)
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
	peer := transport.NewPeer(fixtures.SomePublicIdentity(), nil)
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

type TestWantsProcess struct {
	WantsProcess *replication.WantsProcess
	BlobStorage  *mocks.BlobStorageMock
	HasHandler   *hasHandlerMock
}

func newTestWantsProcess() TestWantsProcess {
	wantListStorage := newWantListStorageMock()
	hasHandler := newHasHandlerMock()
	blobStorage := mocks.NewBlobStorageMock()
	logger := logging.NewDevNullLogger()

	process := replication.NewWantsProcess(
		wantListStorage,
		blobStorage,
		hasHandler,
		logger,
	)

	return TestWantsProcess{
		WantsProcess: process,
		BlobStorage:  blobStorage,
		HasHandler:   hasHandler,
	}
}
