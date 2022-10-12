package ebt_test

import (
	"context"
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

func TestReplicator_ReplicateWaitsForSessionIfConnectionInitiatedByRemote(t *testing.T) {
	testCases := []struct {
		Name                           string
		ConnectionWasInitiatedByRemote bool
		ExpectedWaitForSessionCall     bool
		ExpectedOpenSessionCall        bool
		ExpectedRequest                bool
		ExpectedRunnerCall             bool
	}{
		{
			Name:                           "not_initiated_by_remote",
			ConnectionWasInitiatedByRemote: true,
			ExpectedWaitForSessionCall:     true,
			ExpectedOpenSessionCall:        false,
			ExpectedRequest:                false,
			ExpectedRunnerCall:             false,
		},
		{
			Name:                           "initiated_by_remote",
			ConnectionWasInitiatedByRemote: false,
			ExpectedWaitForSessionCall:     false,
			ExpectedOpenSessionCall:        true,
			ExpectedRequest:                true,
			ExpectedRunnerCall:             true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tr := newTestReplicator(t)

			connectionId := fixtures.SomeConnectionId()
			ctx := fixtures.TestContext(t)
			ctx = rpc.PutConnectionIdInContext(ctx, connectionId)
			conn := newConnectionMock(testCase.ConnectionWasInitiatedByRemote)
			peer := transport.NewPeer(fixtures.SomePublicIdentity(), conn)

			tr.Tracker.WaitForSessionResult = true

			err := tr.Replicator.Replicate(ctx, peer)
			require.NoError(t, err)

			if testCase.ExpectedWaitForSessionCall {
				require.Equal(t,
					[]rpc.ConnectionId{
						connectionId,
					},
					tr.Tracker.WaitForSessionCalls,
				)
			} else {
				require.Empty(t, tr.Tracker.WaitForSessionCalls)
			}

			if testCase.ExpectedOpenSessionCall {
				require.Equal(t,
					[]rpc.ConnectionId{
						connectionId,
					},
					tr.Tracker.OpenSessionCalls,
				)

				require.Equal(t,
					[]rpc.ConnectionId{
						connectionId,
					},
					tr.Tracker.OpenSessionDoneCalls,
				)
			} else {
				require.Empty(t, tr.Tracker.OpenSessionCalls)
				require.Empty(t, tr.Tracker.OpenSessionDoneCalls)
			}

			if testCase.ExpectedRequest {
				require.Equal(t,
					[]*rpc.Request{
						rpc.MustNewRequest(messages.EbtReplicateProcedure.Name(), messages.EbtReplicateProcedure.Typ(), []byte(`[{"version":3,"format":"classic"}]`)),
					},
					conn.PerformRequestCalls,
				)
			} else {
				require.Empty(t, conn.PerformRequestCalls)
			}

			if testCase.ExpectedRunnerCall {
				require.Equal(t, 1, tr.Runner.HandleStreamCallsCount)
			} else {
				require.Equal(t, 0, tr.Runner.HandleStreamCallsCount)
			}
		})
	}
}

func TestReplicator_ReplicateReturnsAnErrorAndDoesNotWaitIfOpenSessionReturnsAnError(t *testing.T) {
	tr := newTestReplicator(t)

	connectionId := fixtures.SomeConnectionId()
	ctx := fixtures.TestContext(t)
	ctx = rpc.PutConnectionIdInContext(ctx, connectionId)
	conn := newConnectionMock(false)
	peer := transport.NewPeer(fixtures.SomePublicIdentity(), conn)

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
	peer := transport.NewPeer(fixtures.SomePublicIdentity(), conn)

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
	Tracker    *trackerMock
	Runner     *runnerMock
	Replicator ebt.Replicator
}

func newTestReplicator(t *testing.T) testReplicator {
	tracker := newTrackerMock()
	runner := newRunnerMock()
	logger := fixtures.TestLogger(t)
	replicator := ebt.NewReplicator(tracker, runner, logger)

	return testReplicator{
		Tracker:    tracker,
		Runner:     runner,
		Replicator: replicator,
	}
}

type connectionMock struct {
	wasInitiatedByRemote bool
	PerformRequestCalls  []*rpc.Request
}

func newConnectionMock(wasInitiatedByRemote bool) *connectionMock {
	return &connectionMock{wasInitiatedByRemote: wasInitiatedByRemote}
}

func (c *connectionMock) PerformRequest(ctx context.Context, req *rpc.Request) (*rpc.ResponseStream, error) {
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

func (t *trackerMock) WaitForSession(ctx context.Context, id rpc.ConnectionId, waitTime time.Duration) bool {
	t.WaitForSessionCalls = append(t.WaitForSessionCalls, id)
	return t.WaitForSessionResult
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
