package tests

import (
	"context"
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

	tr.ContactsRepository.GetWantedFeedsReturnValue = replication.NewWantedFeeds([]replication.Contact{
		{
			Who:       fixtures.SomeRefFeed(),
			Hops:      graph.MustNewHops(fixtures.SomePositiveInt()),
			FeedState: replication.NewEmptyFeedState(),
		},
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
				Err:   rpc.ErrRemoteError,
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
