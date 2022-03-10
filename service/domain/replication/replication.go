package replication

import (
	"context"
	"strconv"

	message2 "github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/network"
	"github.com/planetary-social/go-ssb/service/domain/network/rpc/messages"
	"github.com/planetary-social/go-ssb/service/domain/refs"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
)

type ReplicateFeedTask struct {
	Id    refs.Feed
	State FeedState
	Ctx   context.Context
}

type ReplicationManager interface {
	// GetFeedsToReplicate returns a channel on which replication tasks are received. The channel stays open as long
	// as the passed context isn't cancelled. Cancelling the context cancels all child contexts in the received tasks.
	GetFeedsToReplicate(ctx context.Context) <-chan ReplicateFeedTask
}

type RawMessageHandler interface {
	Handle(msg message2.RawMessage) error
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

func (r GossipReplicator) Replicate(ctx context.Context, peer network.Peer) error {
	feedsToReplicateCh := r.manager.GetFeedsToReplicate(ctx)

	for feed := range feedsToReplicateCh {
		r.replicateFeedTask(peer, feed)
	}

	return nil
}

func (r GossipReplicator) replicateFeedTask(peer network.Peer, feed ReplicateFeedTask) {
	logger := r.logger.WithField("feed", feed.Id.String()).New("replication task")

	logger.WithField("state", feed.State).Debug("starting")

	n, err := r.replicateFeed(peer, feed)
	if err != nil {
		logger.WithField("n", n).WithError(err).Debug("failed")
		return
	}

	logger.WithField("n", n).Debug("finished")
}

func (r GossipReplicator) replicateFeed(peer network.Peer, feed ReplicateFeedTask) (int, error) {
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

		rawMsg := message2.NewRawMessage(response.Value.Bytes())

		if err := r.handler.Handle(rawMsg); err != nil {
			return counter, errors.Wrap(err, "could not process the raw message")
		}

		counter++
	}

	return counter, nil
}

func (r GossipReplicator) newCreateHistoryStreamArguments(id refs.Feed, state FeedState) (messages.CreateHistoryStreamArguments, error) {
	var seq *message2.Sequence
	if sequence, hasAnyMessages := state.Sequence(); hasAnyMessages {
		seq = &sequence
	}

	//tr := true // todo wtf
	fa := false
	limit := 100
	return messages.NewCreateHistoryStreamArguments(id, seq, &limit, &fa, nil, &fa)
}

type FeedState struct {
	sequence *message2.Sequence
}

func NewEmptyFeedState() FeedState {
	return FeedState{}
}

func NewFeedState(sequence message2.Sequence) FeedState {
	return FeedState{
		sequence: &sequence,
	}
}

func (s FeedState) Sequence() (message2.Sequence, bool) {
	if s.sequence != nil {
		return *s.sequence, true
	}
	return message2.Sequence{}, false
}

func (s FeedState) String() string {
	if s.sequence != nil {
		return strconv.Itoa(s.sequence.Int())
	}
	return "empty"
}