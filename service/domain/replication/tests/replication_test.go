package tests

import (
	"context"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestReplicationFallsBackToCreateHistoryStream(t *testing.T) {
	tr, err := BuildTestReplication(t)
	require.NoError(t, err)

	ctx := fixtures.TestContext(t)
	ctx = rpc.PutConnectionIdInContext(ctx, fixtures.SomeConnectionId())

	conn := mocks.NewConnectionMock(ctx)
	conn.SetWasInitiatedByRemote(false)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), conn)

	tr.ContactsRepository.GetWantedFeedsReturnValue = replication.MustNewWantedFeeds([]replication.Contact{
		replication.MustNewContact(
			fixtures.SomeRefFeed(),
			graph.MustNewHops(fixtures.SomePositiveInt()),
			replication.NewEmptyFeedState(),
		),
	}, nil)

	var requests []*rpc.Request
	var requestsLock sync.Mutex

	conn.Mock(func(req *rpc.Request) []rpc.ResponseWithError {
		requestsLock.Lock()
		defer requestsLock.Unlock()

		t.Log("handler received: ", req.Name().String())
		requests = append(requests, req)

		return []rpc.ResponseWithError{
			{
				Value: nil,
				Err:   rpc.NewRemoteError(nil),
			},
		}
	})

	replicateCtx, replicateCancel := context.WithCancel(ctx)
	defer replicateCancel()

	errCh := make(chan error)
	go func() {
		errCh <- tr.Negotiator.Replicate(replicateCtx, peer)
	}()

	require.Eventually(t, func() bool {
		requestsLock.Lock()
		defer requestsLock.Unlock()

		if len(requests) != 2 {
			return false
		}

		if !requests[0].Name().Equal(messages.EbtReplicateProcedure.Name()) {
			return false
		}

		if !requests[1].Name().Equal(messages.CreateHistoryStreamProcedure.Name()) {
			return false
		}

		return true
	}, 1*time.Second, 10*time.Millisecond)

	replicateCancel()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout")
	}
}

func TestReplicationPullsOwnStreamUsingCreateHistoryStreamEvenIfEpidemicBroadcastTreesAreUsed(t *testing.T) {
	tr, err := BuildTestReplication(t)
	require.NoError(t, err)

	ctx := fixtures.TestContext(t)
	ctx = rpc.PutConnectionIdInContext(ctx, fixtures.SomeConnectionId())

	conn := mocks.NewConnectionMock(ctx)
	conn.SetWasInitiatedByRemote(false)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), conn)

	localFeed := fixtures.SomeRefFeed()

	tr.ContactsRepository.GetWantedFeedsReturnValue = replication.MustNewWantedFeeds([]replication.Contact{
		replication.MustNewContact(
			localFeed,
			graph.MustNewHops(0),
			replication.NewEmptyFeedState(),
		),
		replication.MustNewContact(
			fixtures.SomeRefFeed(),
			graph.MustNewHops(fixtures.SomePositiveInt()),
			replication.NewEmptyFeedState(),
		),
	}, nil)

	var requests []*rpc.Request
	var requestsLock sync.Mutex

	replicateCtx, replicateCancel := context.WithCancel(ctx)
	defer replicateCancel()

	conn.Mock(func(req *rpc.Request) []rpc.ResponseWithError {
		requestsLock.Lock()
		requests = append(requests, req)
		requestsLock.Unlock()

		<-replicateCtx.Done()

		return []rpc.ResponseWithError{
			{
				Value: nil,
				Err:   rpc.NewRemoteError([]byte(replicateCtx.Err().Error())),
			},
		}
	})

	errCh := make(chan error)
	go func() {
		errCh <- tr.Negotiator.Replicate(replicateCtx, peer)
	}()

	require.Eventually(t, func() bool {
		requestsLock.Lock()
		defer requestsLock.Unlock()

		sort.Slice(requests, func(i, j int) bool {
			return requests[i].Name().String() < requests[j].Name().String()

		})

		if len(requests) != 2 {
			return false
		}

		ebtRequest := requests[1]
		chsRequest := requests[0]

		if !ebtRequest.Name().Equal(messages.EbtReplicateProcedure.Name()) {
			return false
		}

		if !chsRequest.Name().Equal(messages.CreateHistoryStreamProcedure.Name()) {
			return false
		}

		if !strings.Contains(string(chsRequest.Arguments()), localFeed.String()) {
			return false
		}

		return true
	}, 1*time.Second, 10*time.Millisecond)

	replicateCancel()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout")
	}
}
