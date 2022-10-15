package replication_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	domainmocks "github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
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

	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), domainmocks.NewConnectionMock(ctx))

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
			return len(m.HasHandler.OnHasReceivedCalls()) == 1
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
	HasHandler      *hasHandlerMock
}

func newTestManager(t *testing.T) testManager {
	wantListStorage := newWantListStorageMock()
	blobStorage := mocks.NewBlobStorageMock()
	hasHandler := newHasHandlerMock()
	logger := logging.NewDevNullLogger()

	manager := replication.NewManager(
		wantListStorage,
		blobStorage,
		hasHandler,
		logger,
	)

	return testManager{
		Manager:         manager,
		WantListStorage: wantListStorage,
		HasHandler:      hasHandler,
	}
}

type wantListStorageMock struct {
	WantList []blobs.WantedBlob
}

func newWantListStorageMock() *wantListStorageMock {
	return &wantListStorageMock{}
}

func (w wantListStorageMock) List() (blobs.WantList, error) {
	return blobs.NewWantList(w.WantList)
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
