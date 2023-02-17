package replication

import (
	"context"
	"sync"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type ManagedWantsProcessFactory interface {
	NewWantsProcess() ManagedWantsProcess
}

type ManagedWantsProcess interface {
	AddIncoming(ctx context.Context, ch chan<- messages.BlobWithSizeOrWantDistance)
	AddOutgoing(ctx context.Context, ch <-chan messages.BlobWithSizeOrWantDistance, peer transport.Peer)
}

type Manager struct {
	factory ManagedWantsProcessFactory
	logger  logging.Logger

	// todo cleanup processes
	processes     map[rpc.ConnectionId]ManagedWantsProcess
	processesLock sync.Mutex
}

func NewManager(
	factory ManagedWantsProcessFactory,
	logger logging.Logger,
) *Manager {
	return &Manager{
		factory: factory,
		logger:  logger.New("blobs_replication_manager"),

		processes: make(map[rpc.ConnectionId]ManagedWantsProcess),
	}
}

func (m *Manager) HandleIncomingCreateWantsRequest(ctx context.Context) (<-chan messages.BlobWithSizeOrWantDistance, error) {
	connectionId, ok := rpc.GetConnectionIdFromContext(ctx)
	if !ok {
		return nil, errors.New("connection id not found in context")
	}

	m.processesLock.Lock()
	defer m.processesLock.Unlock()

	ch := make(chan messages.BlobWithSizeOrWantDistance)
	m.getOrCreateProcess(connectionId).AddIncoming(ctx, ch)
	return ch, nil
}

func (m *Manager) HandleOutgoingCreateWantsRequest(ctx context.Context, ch <-chan messages.BlobWithSizeOrWantDistance, peer transport.Peer) error {
	connectionId, ok := rpc.GetConnectionIdFromContext(ctx)
	if !ok {
		return errors.New("connection id not found in context")
	}

	m.processesLock.Lock()
	defer m.processesLock.Unlock()

	m.getOrCreateProcess(connectionId).AddOutgoing(ctx, ch, peer)
	return nil
}

func (m *Manager) getOrCreateProcess(id rpc.ConnectionId) ManagedWantsProcess {
	v, ok := m.processes[id]
	if !ok {
		v = m.factory.NewWantsProcess()
		m.processes[id] = v
	}
	return v
}
