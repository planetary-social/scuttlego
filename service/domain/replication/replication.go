package replication

import (
	"context"
	"strconv"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/messages"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/transport"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
)

const (
	numWorkers = 10
	limit      = 1000
)

type TaskResult struct {
	s string
}

var (
	TaskResultDoesNotHaveMoreMessages = TaskResult{"does_not_have_more_messages"}
	TaskResultHasMoreMessages         = TaskResult{"has_more_messages"}
	TaskResultFailed                  = TaskResult{"failed"}
	TaskResultDidNotStart             = TaskResult{"did_not_start"}
)

type TaskCompletedFn func(result TaskResult)

type ReplicateFeedTask struct {
	Id       refs.Feed
	State    FeedState
	Ctx      context.Context
	Complete TaskCompletedFn
}

type ReplicationManager interface {
	// GetFeedsToReplicate returns a channel on which replication tasks are
	// received. The channel stays open as long as the passed context isn't
	// cancelled. Cancelling the context cancels all child contexts in the
	// received tasks.
	GetFeedsToReplicate(ctx context.Context, remote identity.Public) <-chan ReplicateFeedTask
}

type RawMessageHandler interface {
	Handle(msg message.RawMessage) error
}

type GossipReplicator struct {
	manager ReplicationManager
	handler RawMessageHandler
	logger  logging.Logger
}

func NewGossipReplicator(manager ReplicationManager, handler RawMessageHandler, logger logging.Logger) (*GossipReplicator, error) {
	return &GossipReplicator{
		manager: manager,
		handler: handler,
		logger:  logger,
	}, nil
}

func (r GossipReplicator) Replicate(ctx context.Context, peer transport.Peer) error {
	feedsToReplicateCh := r.manager.GetFeedsToReplicate(ctx, peer.Identity())
	r.startWorkers(peer, feedsToReplicateCh)
	<-ctx.Done()

	return nil
}

func (r GossipReplicator) startWorkers(peer transport.Peer, ch <-chan ReplicateFeedTask) {
	for i := 0; i < numWorkers; i++ {
		go r.worker(peer, ch)
	}
}

func (r GossipReplicator) worker(peer transport.Peer, ch <-chan ReplicateFeedTask) {
	for task := range ch {
		r.replicateFeedTask(peer, task)
	}
}

func (r GossipReplicator) replicateFeedTask(peer transport.Peer, task ReplicateFeedTask) {
	logger := r.logger.
		WithField("feed", task.Id.String()).
		WithField("state", task.State).
		New("replication task")

	n, err := r.replicateFeed(peer, task)
	logger.WithField("n", n).WithError(err).Debug("finished")
	if err != nil && !errors.Is(err, rpc.ErrEndOrErr) {
		task.Complete(TaskResultFailed)
		return
	}

	if n < limit {
		task.Complete(TaskResultDoesNotHaveMoreMessages)
	} else {
		task.Complete(TaskResultHasMoreMessages)
	}
}

func (r GossipReplicator) replicateFeed(peer transport.Peer, feed ReplicateFeedTask) (int, error) {
	ctx, cancel := context.WithCancel(feed.Ctx)
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

		rawMsg := message.NewRawMessage(response.Value.Bytes())

		if err := r.handler.Handle(rawMsg); err != nil {
			return counter, errors.Wrap(err, "could not process the raw message")
		}

		counter++
	}

	return counter, nil
}

func (r GossipReplicator) newCreateHistoryStreamArguments(id refs.Feed, state FeedState) (messages.CreateHistoryStreamArguments, error) {
	var seq *message.Sequence
	if sequence, hasAnyMessages := state.Sequence(); hasAnyMessages {
		seq = &sequence
	}

	fa := false
	limit := limit
	return messages.NewCreateHistoryStreamArguments(id, seq, &limit, &fa, nil, &fa)
}

type FeedState struct {
	sequence *message.Sequence
}

func NewEmptyFeedState() FeedState {
	return FeedState{}
}

func NewFeedState(sequence message.Sequence) FeedState {
	return FeedState{
		sequence: &sequence,
	}
}

func (s FeedState) Sequence() (message.Sequence, bool) {
	if s.sequence != nil {
		return *s.sequence, true
	}
	return message.Sequence{}, false
}

func (s FeedState) String() string {
	if s.sequence != nil {
		return strconv.Itoa(s.sequence.Int())
	}
	return "empty"
}
