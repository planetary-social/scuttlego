package replication_test

import (
	"context"
	"testing"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/internal/mocks"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestManager_HandleIncomingCreateWantsRequest_CreatesProcessAndCallsAddIncoming(t *testing.T) {
	ts := newTestManager(t)

	ctx := newConnectionContext(t)

	_, err := ts.Manager.HandleIncomingCreateWantsRequest(ctx)
	require.NoError(t, err)

	require.Len(t, ts.Factory.CreatedProcesses, 1)
	require.Equal(t, 1, ts.Factory.CreatedProcesses[0].AddIncomingCallsCount)
}

func TestManager_HandleOutgoingCreateWantsRequest_CreatesProcessAndCallsAddOutgoing(t *testing.T) {
	ts := newTestManager(t)

	ctx := newConnectionContext(t)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), mocks.NewConnectionMock(ctx))

	err := ts.Manager.HandleOutgoingCreateWantsRequest(ctx, nil, peer)
	require.NoError(t, err)

	require.Len(t, ts.Factory.CreatedProcesses, 1)
	require.Equal(t, 1, ts.Factory.CreatedProcesses[0].AddOutgoingCallsCount)
}

func newConnectionContext(t *testing.T) context.Context {
	ctx := fixtures.TestContext(t)
	return rpc.PutConnectionIdInContext(ctx, fixtures.SomeConnectionId())
}

type testManager struct {
	Manager *replication.Manager
	Factory *managedWantsProcessFactoryMock
}

func newTestManager(t *testing.T) testManager {
	factory := newManagedWantsProcessFactoryMock()
	logger := logging.NewDevNullLogger()

	manager := replication.NewManager(
		factory,
		logger,
	)

	return testManager{
		Manager: manager,
		Factory: factory,
	}
}

type managedWantsProcessFactoryMock struct {
	CreatedProcesses []*managedWantsProcessMock
}

func newManagedWantsProcessFactoryMock() *managedWantsProcessFactoryMock {
	return &managedWantsProcessFactoryMock{}
}

func (m *managedWantsProcessFactoryMock) NewWantsProcess() replication.ManagedWantsProcess {
	v := newManagedWantsProcessMock()
	m.CreatedProcesses = append(m.CreatedProcesses, v)
	return v
}

type managedWantsProcessMock struct {
	AddIncomingCallsCount int
	AddOutgoingCallsCount int
}

func newManagedWantsProcessMock() *managedWantsProcessMock {
	return &managedWantsProcessMock{}
}

func (m *managedWantsProcessMock) AddIncoming(ctx context.Context, ch chan<- messages.BlobWithSizeOrWantDistance) {
	m.AddIncomingCallsCount++
}

func (m *managedWantsProcessMock) AddOutgoing(ctx context.Context, ch <-chan messages.BlobWithSizeOrWantDistance, peer transport.Peer) {
	m.AddOutgoingCallsCount++
}
