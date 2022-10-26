package tests

import (
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

	tr.ContactsRepository.GetContactsReturnValue = []replication.Contact{
		{
			Who:       fixtures.SomeRefFeed(),
			Hops:      graph.MustNewHops(fixtures.SomePositiveInt()),
			FeedState: replication.NewEmptyFeedState(),
		},
	}

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

	go func() {
		err = tr.Negotiator.Replicate(ctx, peer)
		t.Log("replicate error", err)
	}()

	require.Eventually(t, func() bool {
		requestsLock.Lock()
		defer requestsLock.Unlock()

		var calledEbt, calledChs bool
		for _, req := range requests {
			if req.Name().Equal(messages.CreateHistoryStreamProcedure.Name()) {
				calledChs = true
			}
			if req.Name().Equal(messages.EbtReplicateProcedure.Name()) {
				calledEbt = true
			}
		}

		return calledEbt && calledChs
	}, 1*time.Second, 10*time.Millisecond)
}
