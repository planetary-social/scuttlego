package ebt_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestReplicator_ReplicateCallsWaitForSessionIfConnectionWasInitiatedByRemote(t *testing.T) {
	tr := newTestReplicator(t)

	connectionId := fixtures.SomeConnectionId()
	ctx := fixtures.TestContext(t)
	ctx = rpc.PutConnectionIdInContext(ctx, connectionId)
	connThatWasInitiatedByRemote := newConnectionMock(true)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), connThatWasInitiatedByRemote)

	tr.Tracker.WaitForSessionResult = true

	err := tr.Replicator.Replicate(ctx, peer)
	require.NoError(t, err)

	require.Equal(t, []rpc.ConnectionId{connectionId}, tr.Tracker.WaitForSessionCalls)
	require.Empty(t, tr.Tracker.OpenSessionCalls)
	require.Empty(t, tr.Tracker.OpenSessionDoneCalls)
	require.Empty(t, connThatWasInitiatedByRemote.PerformRequestCalls)
	require.Equal(t, 0, tr.Runner.HandleStreamCallsCount)
}

func TestReplicator_ReplicateInitiatesTheSessionIfConnectionWasNotInitiatedByRemote(t *testing.T) {
	tr := newTestReplicator(t)

	connectionId := fixtures.SomeConnectionId()
	ctx := fixtures.TestContext(t)
	ctx = rpc.PutConnectionIdInContext(ctx, connectionId)
	connectionThatWasNotInitiatedByRemote := newConnectionMock(false)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), connectionThatWasNotInitiatedByRemote)

	err := tr.Replicator.Replicate(ctx, peer)
	require.NoError(t, err)

	require.Empty(t, tr.Tracker.WaitForSessionCalls)
	require.Equal(t, []rpc.ConnectionId{connectionId}, tr.Tracker.OpenSessionCalls)
	require.Equal(t, []rpc.ConnectionId{connectionId}, tr.Tracker.OpenSessionDoneCalls)
	require.Equal(t,
		[]*rpc.Request{
			rpc.MustNewRequest(messages.EbtReplicateProcedure.Name(), messages.EbtReplicateProcedure.Typ(), []byte(`[{"version":3,"format":"classic"}]`)),
		},
		connectionThatWasNotInitiatedByRemote.PerformRequestCalls,
	)
	require.Equal(t, 1, tr.Runner.HandleStreamCallsCount)
}

func TestReplicator_ReplicateCallsReplicateSelfIfConnectionWasInitiatedByRemote(t *testing.T) {
	tr := newTestReplicator(t)

	connectionId := fixtures.SomeConnectionId()
	ctx := fixtures.TestContext(t)
	ctx = rpc.PutConnectionIdInContext(ctx, connectionId)
	connThatWasInitiatedByRemote := newConnectionMock(true)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), connThatWasInitiatedByRemote)

	tr.Tracker.WaitForSessionResult = true

	err := tr.Replicator.Replicate(ctx, peer)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return len(tr.SelfReplicator.ReplicateSelfCalls()) > 0
	}, 1*time.Second, 10*time.Millisecond)
}

func TestReplicator_ReplicateCallsReplicateSelfIfConnectionWasNotInitiatedByRemote(t *testing.T) {
	tr := newTestReplicator(t)

	connectionId := fixtures.SomeConnectionId()
	ctx := fixtures.TestContext(t)
	ctx = rpc.PutConnectionIdInContext(ctx, connectionId)
	connectionThatWasNotInitiatedByRemote := newConnectionMock(false)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), connectionThatWasNotInitiatedByRemote)

	err := tr.Replicator.Replicate(ctx, peer)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return len(tr.SelfReplicator.ReplicateSelfCalls()) > 0
	}, 1*time.Second, 10*time.Millisecond)
}

func TestReplicator_ReplicateReturnsAnErrorAndDoesNotWaitIfOpenSessionReturnsAnError(t *testing.T) {
	tr := newTestReplicator(t)

	connectionId := fixtures.SomeConnectionId()
	ctx := fixtures.TestContext(t)
	ctx = rpc.PutConnectionIdInContext(ctx, connectionId)
	conn := newConnectionMock(false)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), conn)

	tr.Tracker.OpenSessionError = fixtures.SomeError()

	err := tr.Replicator.Replicate(ctx, peer)
	require.ErrorIs(t, err, tr.Tracker.OpenSessionError)

	require.Equal(t, []rpc.ConnectionId{connectionId}, tr.Tracker.OpenSessionCalls)
	require.Empty(t, tr.Tracker.OpenSessionDoneCalls)
	require.Empty(t, tr.Tracker.WaitForSessionCalls)
	require.Empty(t, conn.PerformRequestCalls)
	require.Equal(t, 0, tr.Runner.HandleStreamCallsCount)
}

func TestReplicator_ReplicateReturnsErrPeerDoesNotSupportEbtIfRemoteNeverOpensASession(t *testing.T) {
	tr := newTestReplicator(t)

	connectionId := fixtures.SomeConnectionId()
	ctx := fixtures.TestContext(t)
	ctx = rpc.PutConnectionIdInContext(ctx, connectionId)
	conn := newConnectionMock(true)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), conn)

	tr.Tracker.WaitForSessionResult = false

	err := tr.Replicator.Replicate(ctx, peer)
	require.ErrorIs(t, err, replication.ErrPeerDoesNotSupportEBT)

	require.Empty(t, tr.Tracker.OpenSessionCalls)
	require.Empty(t, tr.Tracker.OpenSessionDoneCalls)
	require.NotEmpty(t, tr.Tracker.WaitForSessionCalls)
	require.Empty(t, conn.PerformRequestCalls)
	require.Equal(t, 0, tr.Runner.HandleStreamCallsCount)
}

type testReplicator struct {
	Tracker        *trackerMock
	Runner         *runnerMock
	SelfReplicator *selfCreateHistoryStreamReplicatorMock
	Replicator     ebt.Replicator
}

func newTestReplicator(t *testing.T) testReplicator {
	tracker := newTrackerMock()
	runner := newRunnerMock()
	logger := fixtures.TestLogger(t)
	selfReplicator := newSelfCreateHistoryStreamReplicatorMock()
	replicator := ebt.NewReplicator(tracker, runner, selfReplicator, logger)

	return testReplicator{
		Tracker:        tracker,
		Runner:         runner,
		SelfReplicator: selfReplicator,
		Replicator:     replicator,
	}
}

type connectionMock struct {
	wasInitiatedByRemote bool
	PerformRequestCalls  []*rpc.Request
}

func newConnectionMock(wasInitiatedByRemote bool) *connectionMock {
	return &connectionMock{wasInitiatedByRemote: wasInitiatedByRemote}
}

func (c *connectionMock) PerformRequest(ctx context.Context, req *rpc.Request) (rpc.ResponseStream, error) {
	c.PerformRequestCalls = append(c.PerformRequestCalls, req)
	return nil, nil
}

func (c *connectionMock) Context() context.Context {
	panic("implement me")
}

func (c *connectionMock) Close() error {
	panic("implement me")
}

func (c connectionMock) WasInitiatedByRemote() bool {
	return c.wasInitiatedByRemote
}

type trackerMock struct {
	WaitForSessionCalls  []rpc.ConnectionId
	OpenSessionCalls     []rpc.ConnectionId
	OpenSessionDoneCalls []rpc.ConnectionId
	OpenSessionError     error
	WaitForSessionResult bool
}

func newTrackerMock() *trackerMock {
	return &trackerMock{}
}

func (t *trackerMock) OpenSession(id rpc.ConnectionId) (ebt.SessionEndedFn, error) {
	t.OpenSessionCalls = append(t.OpenSessionCalls, id)
	return func() {
		t.OpenSessionDoneCalls = append(t.OpenSessionDoneCalls, id)
	}, t.OpenSessionError
}

func (t *trackerMock) WaitForSession(ctx context.Context, id rpc.ConnectionId, waitTime time.Duration) (bool, error) {
	t.WaitForSessionCalls = append(t.WaitForSessionCalls, id)
	return t.WaitForSessionResult, nil
}

type runnerMock struct {
	HandleStreamCallsCount int
}

func newRunnerMock() *runnerMock {
	return &runnerMock{}
}

func (r *runnerMock) HandleStream(ctx context.Context, stream ebt.Stream) error {
	r.HandleStreamCallsCount++
	return nil
}

type replicateSelfCall struct {
	Peer transport.Peer
}

type selfCreateHistoryStreamReplicatorMock struct {
	replicateSelfCalls     []replicateSelfCall
	replicateSelfCallsLock sync.Mutex
}

func newSelfCreateHistoryStreamReplicatorMock() *selfCreateHistoryStreamReplicatorMock {
	return &selfCreateHistoryStreamReplicatorMock{}
}

func (s *selfCreateHistoryStreamReplicatorMock) ReplicateSelf(ctx context.Context, peer transport.Peer) error {
	s.replicateSelfCallsLock.Lock()
	s.replicateSelfCalls = append(s.replicateSelfCalls, replicateSelfCall{Peer: peer})
	s.replicateSelfCallsLock.Unlock()

	<-ctx.Done()
	return ctx.Err()
}

func (s *selfCreateHistoryStreamReplicatorMock) ReplicateSelfCalls() []replicateSelfCall {
	s.replicateSelfCallsLock.Lock()
	defer s.replicateSelfCallsLock.Unlock()
	tmp := make([]replicateSelfCall, len(s.replicateSelfCalls))
	copy(tmp, s.replicateSelfCalls)
	return tmp
}
