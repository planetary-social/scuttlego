package gossip

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

const (
	// How many tasks will be executed at the same time for a peer. Effectively
	// the number of feeds that are concurrently replicated per peer.
	numWorkers = 10

	// How many messages to ask for whenever a replication task is received. The
	// number shouldn't be too large so that the tasks are not very long-lived
	// and not to small to reduce the overhead related to sending new RPC
	// requests too often.
	limit = 1000
)

type GossipReplicator struct {
	manager ReplicationManager
	handler replication.RawMessageHandler
	logger  logging.Logger
}

func NewGossipReplicator(manager ReplicationManager, handler replication.RawMessageHandler, logger logging.Logger) (*GossipReplicator, error) {
	return &GossipReplicator{
		manager: manager,
		handler: handler,
		logger:  logger.New("gossip_replicator"),
	}, nil
}

func (r GossipReplicator) Replicate(ctx context.Context, peer transport.Peer) error {
	feedsToReplicateCh := r.manager.GetFeedsToReplicate(ctx, peer.Identity())
	r.startWorkers(ctx, peer, feedsToReplicateCh)
	<-ctx.Done()

	return nil
}

func (r GossipReplicator) ReplicateSelf(ctx context.Context, peer transport.Peer) error {
	feedsToReplicateCh := r.manager.GetFeedsToReplicateSelf(ctx, peer.Identity())
	r.startWorkers(ctx, peer, feedsToReplicateCh)
	<-ctx.Done()

	return nil
}

func (r GossipReplicator) startWorkers(ctx context.Context, peer transport.Peer, ch <-chan ReplicateFeedTask) {
	for i := 0; i < numWorkers; i++ {
		go r.worker(ctx, peer, ch)
	}
}

func (r GossipReplicator) worker(ctx context.Context, peer transport.Peer, ch <-chan ReplicateFeedTask) {
	for task := range ch {
		r.replicateFeedTask(ctx, peer, task)
	}
}

func (r GossipReplicator) replicateFeedTask(ctx context.Context, peer transport.Peer, task ReplicateFeedTask) {
	logger := r.logger.
		WithField("peer", peer.Identity().String()).
		WithField("feed", task.Id.String()).
		WithField("state", task.State).
		New("task")

	logger.Trace("starting")

	n, err := r.replicateFeed(ctx, peer, task)
	if err != nil && !errors.Is(err, rpc.ErrRemoteEnd) {
		logger.WithField("received_messages", n).WithError(err).Error("failed")
		task.OnComplete(TaskResultFailed)
		return
	}

	logger.WithField("received_messages", n).Trace("finished")

	if n < limit {
		task.OnComplete(TaskResultDoesNotHaveMoreMessages)
	} else {
		task.OnComplete(TaskResultHasMoreMessages)
	}
}

func (r GossipReplicator) replicateFeed(ctx context.Context, peer transport.Peer, feed ReplicateFeedTask) (int, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	arguments, err := r.newCreateHistoryStreamArguments(feed.Id, feed.State)
	if err != nil {
		return 0, errors.Wrap(err, "could not create history stream arguments")
	}

	request, err := messages.NewCreateHistoryStream(arguments)
	if err != nil {
		return 0, errors.Wrap(err, "could not create a request")
	}

	rs, err := peer.Conn().PerformRequest(ctx, request)
	if err != nil {
		return 0, errors.Wrap(err, "could not perform a request")
	}

	counter := 0
	for response := range rs.Channel() {
		if err := response.Err; err != nil {
			return counter, errors.Wrap(err, "response stream error")
		}

		rawMsg, err := message.NewRawMessage(response.Value.Bytes())
		if err != nil {
			return counter, errors.Wrap(err, "could not create a raw message")
		}

		if err := r.handler.Handle(peer.Identity(), rawMsg); err != nil {
			return counter, errors.Wrap(err, "could not process the raw message")
		}

		counter++
	}

	return counter, nil
}

func (r GossipReplicator) newCreateHistoryStreamArguments(id refs.Feed, state replication.FeedState) (messages.CreateHistoryStreamArguments, error) {
	var seq *message.Sequence
	if sequence, hasAnyMessages := state.Sequence(); hasAnyMessages {
		seq = &sequence
	}

	fa := false
	limit := limit
	return messages.NewCreateHistoryStreamArguments(id, seq, &limit, &fa, nil, &fa)
}
