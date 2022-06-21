package replication_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/blobs"
	"github.com/planetary-social/go-ssb/service/domain/blobs/replication"
	"github.com/planetary-social/go-ssb/service/domain/messages"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/transport"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestManagerSendsLocalWants(t *testing.T) {
	m := newTestManager(t)
	ctx := newConnectionContext(t)

	m.WantListStorage.WantList = []blobs.WantedBlob{
		{
			Id:       fixtures.SomeRefBlob(),
			Distance: fixtures.SomeWantDistance(),
		},
	}

	ch, err := m.Manager.HandleIncomingCreateWantsRequest(ctx)
	require.NoError(t, err)

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

	require.Equal(t, len(m.WantListStorage.WantList), len(localWants))
}

func TestManagerTriggersDownloaderAfterReceivingHas(t *testing.T) {
	m := newTestManager(t)
	ctx := newConnectionContext(t)

	ch := make(chan messages.BlobWithSizeOrWantDistance)

	peer := transport.NewPeer(fixtures.SomePublicIdentity(), nil)

	err := m.Manager.HandleOutgoingCreateWantsRequest(ctx, ch, peer)
	require.NoError(t, err)

	blob, err := messages.NewBlobWithSize(fixtures.SomeRefBlob(), fixtures.SomeSize())
	require.NoError(t, err)

	select {
	case ch <- blob:
		// ok
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}

	require.Eventually(t,
		func() bool {
			return len(m.Downloader.OnHasReceivedCalls()) == 1
		},
		1*time.Second,
		10*time.Millisecond,
	)
}

func newConnectionContext(t *testing.T) context.Context {
	ctx := fixtures.TestContext(t)
	return rpc.PutConnectionIdInContext(ctx, fixtures.SomeConnectionId())
}

type testManager struct {
	Manager *replication.Manager

	WantListStorage *wantListStorageMock
	Downloader      *downloaderMock
}

func newTestManager(t *testing.T) testManager {
	wantListStorage := newWantListStorageMock()
	downloader := newDownloaderMock()
	logger := logging.NewDevNullLogger()

	manager := replication.NewManager(
		wantListStorage,
		downloader,
		logger,
	)

	return testManager{
		Manager:         manager,
		WantListStorage: wantListStorage,
		Downloader:      downloader,
	}
}

type wantListStorageMock struct {
	WantList []blobs.WantedBlob
}

func newWantListStorageMock() *wantListStorageMock {
	return &wantListStorageMock{}
}

func (w wantListStorageMock) GetWantList() (blobs.WantList, error) {
	return blobs.NewWantList(w.WantList)
}

type downloaderMock struct {
	onHasReceivedCalls []onHasReceivedCall
	lock               sync.Mutex
}

func newDownloaderMock() *downloaderMock {
	return &downloaderMock{}
}

func (d *downloaderMock) OnHasReceived(ctx context.Context, peer transport.Peer, blob refs.Blob, size blobs.Size) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.onHasReceivedCalls = append(d.onHasReceivedCalls, onHasReceivedCall{
		Peer: peer,
		Blob: blob,
		Size: size,
	})
}

func (d *downloaderMock) OnHasReceivedCalls() []onHasReceivedCall {
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
