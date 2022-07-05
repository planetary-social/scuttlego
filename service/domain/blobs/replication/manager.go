package replication

import (
	"context"
	"sync"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/messages"
	"github.com/planetary-social/go-ssb/service/domain/transport"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
)

type Manager struct {
	wantListStorage WantListStorage
	blobStorage     BlobSizeRepository
	hasHandler      HasBlobHandler
	logger          logging.Logger

	// todo cleanup processes
	processes map[rpc.ConnectionId]*WantsProcess
	lock      sync.Mutex // guards processes
}

func NewManager(
	wantListStorage WantListStorage,
	blobStorage BlobSizeRepository,
	hasHandler HasBlobHandler,
	logger logging.Logger,
) *Manager {
	return &Manager{
		wantListStorage: wantListStorage,
		blobStorage:     blobStorage,
		hasHandler:      hasHandler,
		processes:       make(map[rpc.ConnectionId]*WantsProcess),
		logger:          logger.New("replication_manager"),
	}
}

func (m *Manager) HandleIncomingCreateWantsRequest(ctx context.Context) (<-chan messages.BlobWithSizeOrWantDistance, error) {
	connectionId, ok := rpc.GetConnectionIdFromContext(ctx)
	if !ok {
		return nil, errors.New("connection id not found in context")
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	ch := make(chan messages.BlobWithSizeOrWantDistance)
	m.getOrCreateProcess(connectionId).AddIncoming(ctx, ch)
	return ch, nil
}

func (m *Manager) HandleOutgoingCreateWantsRequest(ctx context.Context, ch <-chan messages.BlobWithSizeOrWantDistance, peer transport.Peer) error {
	connectionId, ok := rpc.GetConnectionIdFromContext(ctx)
	if !ok {
		return errors.New("connection id not found in context")
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	m.getOrCreateProcess(connectionId).AddOutgoing(ctx, ch, peer)
	return nil
}

func (m *Manager) getOrCreateProcess(id rpc.ConnectionId) *WantsProcess {
	v, ok := m.processes[id]
	if !ok {
		v = NewWantsProcess(
			m.wantListStorage,
			m.blobStorage,
			m.hasHandler,
			m.logger.WithField("connection_id", id),
		)
		m.processes[id] = v
	}
	return v
}
