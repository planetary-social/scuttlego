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
	// How many tasks will be executed at the same time for a peer. Effectively
	// the number of feeds that are concurrently replicated per peer.
	numWorkers = 10

	// How many messages to ask for whenever a replication task is received. The
	// number shouldn't be too large so that the tasks are not very long-lived
	// and not to small to reduce the overhead related to sending new RPC
	// requests too often.
	limit = 1000
)

type TaskResult struct {
	s string
}

var (
	TaskResultDoesNotHaveMoreMessages = TaskResult{"does_not_have_more_messages"}
	TaskResultHasMoreMessages         = TaskResult{"has_more_messages"}
	TaskResultFailed                  = TaskResult{"failed"}

	// TaskResultDidNotStart is used internally by the manager. It should not be
	// used by replicators.
	TaskResultDidNotStart = TaskResult{"did_not_start"}
)

type TaskCompletedFn func(result TaskResult)

type ReplicateFeedTask struct {
	Id    refs.Feed
	State FeedState
	Ctx   context.Context

	OnComplete TaskCompletedFn
}

type ReplicationManager interface {
	// GetFeedsToReplicate returns a channel on which replication tasks are
	// received. The channel stays open as long as the passed context isn't
	// cancelled. Cancelling the context cancels all child contexts in the
	// received tasks. The caller must call the completion function for each
	// task.
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
		WithField("peer", peer.Identity().String()).
		WithField("feed", task.Id.String()).
		WithField("state", task.State).
		New("replication task")

	n, err := r.replicateFeed(peer, task)
	if err != nil && !errors.Is(err, rpc.ErrEndOrErr) {
		logger.WithField("n", n).WithError(err).Debug("failed")
		task.OnComplete(TaskResultFailed)
		return
	}

	logger.WithField("n", n).Debug("finished")

	if n < limit {
		task.OnComplete(TaskResultDoesNotHaveMoreMessages)
	} else {
		task.OnComplete(TaskResultHasMoreMessages)
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

		rawMsg, err := message.NewRawMessage(response.Value.Bytes())
		if err != nil {
			return counter, errors.Wrap(err, "could not create a raw message")
		}

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

// FeedState wraps the sequence number so that both the state of feeds which
// have some messages in them and empty feeds can be represented.
type FeedState struct {
	sequence *message.Sequence
}

// NewEmptyFeedState creates a new feed state which represents an empty feed.
// This is equivalent to the zero value of this type but using this constructor
// improves readability.
func NewEmptyFeedState() FeedState {
	return FeedState{}
}

// NewFeedState creates a new feed state which represents a feed for which at
// least one message is known.
func NewFeedState(sequence message.Sequence) (FeedState, error) {
	if sequence.IsZero() {
		return FeedState{}, errors.New("zero value of sequence")
	}

	return FeedState{
		sequence: &sequence,
	}, nil
}

// Sequence returns the sequence of the last message in the feed. If the feed is
// empty then the sequence is not returned.
func (s FeedState) Sequence() (message.Sequence, bool) {
	if s.sequence != nil {
		return *s.sequence, true
	}
	return message.Sequence{}, false
}

// String is useful for printing this value when logging or debugging. Do not
// use it for other purposes.
func (s FeedState) String() string {
	if s.sequence != nil {
		return strconv.Itoa(s.sequence.Int())
	}
	return "empty"
}
