package replication

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/network"
	"github.com/planetary-social/go-ssb/network/rpc/messages"
	"github.com/planetary-social/go-ssb/refs"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
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

func (r GossipReplicator) Replicate(ctx context.Context, peer network.Peer) error {
	feedsToReplicateCh := r.manager.GetFeedsToReplicate(ctx)

	for feed := range feedsToReplicateCh {
		if err := r.replicateFeedTask(peer, feed); err != nil {
			return errors.Wrap(err, "could not replicate")
		}
	}

	return nil
}

func (r GossipReplicator) replicateFeedTask(peer network.Peer, feed ReplicateFeedTask) error {
	r.logger.WithField("feed", feed.Id.String()).WithField("state", feed.State).Debug("new replicate task")

	n, err := r.replicateFeed(peer, feed)
	if err != nil {
		r.logger.WithField("n", n).WithError(err).Error("could not replicate a feed")
		return errors.Wrap(err, "replication failed")
	}

	r.logger.WithField("n", n).Debug("replicated feed")

	return nil
}

func (r GossipReplicator) replicateFeed(peer network.Peer, feed ReplicateFeedTask) (int, error) {
	arguments, err := r.newCreateHistoryStreamArguments(feed.Id, feed.State)
	if err != nil {
		return 0, errors.Wrap(err, "could not create history stream arguments")
	}

	request, err := messages.NewCreateHistoryStream(arguments)
	if err != nil {
		return 0, errors.Wrap(err, "could not create a request")
	}

	rs, err := peer.Conn().PerformRequest(feed.Ctx, request)
	if err != nil {
		return 0, errors.Wrap(err, "could not perform a request")
	}

	counter := 0
	for response := range rs.Channel() {
		if err := response.Err; err != nil {
			return counter, errors.Wrap(err, "response stream error") // todo what to do if an error is returned
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
	tr := true // todo wtf
	fa := false
	if sequence, hasAnyMessages := state.Sequence(); hasAnyMessages {
		return messages.NewCreateHistoryStreamArguments(id, &sequence, nil, &tr, nil, &fa)
	} else {
		return messages.NewCreateHistoryStreamArguments(id, nil, nil, &tr, nil, &fa)
	}
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
