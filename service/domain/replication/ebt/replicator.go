package ebt

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

const (
	waitForRemoteToStartEbtSessionFor = 5 * time.Second
	ebtReplicateVersion               = 3
)

var (
	ebtReplicateFormat = messages.EbtReplicateFormatClassic
)

type SelfCreateHistoryStreamReplicator interface {
	// ReplicateSelf should keep attempting to perform replication as long as
	// the context isn't closed. Returning an error implies that replication
	// should not restart.
	ReplicateSelf(ctx context.Context, peer transport.Peer) error
}

type Tracker interface {
	OpenSession(id rpc.ConnectionId) (SessionEndedFn, error)

	// WaitForSession waits for the session to be started for the provided
	// amount of time. If the session starts within the provided time window
	// then WaitForSession blocks for as long as the session is running.
	// Returning true signifies that the session existed at any point after
	// calling this function. Error is returned if the context is cancelled.
	WaitForSession(ctx context.Context, id rpc.ConnectionId, waitTime time.Duration) (bool, error)
}

type Runner interface {
	HandleStream(ctx context.Context, stream Stream) error
}

type Replicator struct {
	tracker                           Tracker
	runner                            Runner
	selfCreateHistoryStreamReplicator SelfCreateHistoryStreamReplicator
	logger                            logging.Logger
}

func NewReplicator(
	tracker Tracker,
	runner Runner,
	selfCreateHistoryStreamReplicator SelfCreateHistoryStreamReplicator,
	logger logging.Logger,
) Replicator {
	return Replicator{
		tracker:                           tracker,
		runner:                            runner,
		selfCreateHistoryStreamReplicator: selfCreateHistoryStreamReplicator,
		logger:                            logger,
	}
}

func (r Replicator) Replicate(ctx context.Context, peer transport.Peer) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	connectionId, ok := rpc.GetConnectionIdFromContext(ctx)
	if !ok {
		return errors.New("connection id not found in context")
	}

	logger := r.logger.WithField("peer", peer)

	if !peer.Conn().WasInitiatedByRemote() {
		logger.Debug("initializing an EBT session")

		done, err := r.tracker.OpenSession(connectionId)
		if err != nil {
			return errors.Wrap(err, "failed to mark local session as open")
		}
		defer done()

		rs, err := r.openEbtStream(ctx, peer)
		if err != nil {
			return errors.Wrap(err, "error starting the ebt session")
		}

		go r.replicateSelf(rs.Ctx(), peer)

		return r.runner.HandleStream(rs.Ctx(), NewOutgoingStreamAdapter(peer.Identity(), rs))
	}

	go r.replicateSelf(ctx, peer)

	logger.Debug("waiting for an EBT session")
	ok, err := r.tracker.WaitForSession(ctx, connectionId, waitForRemoteToStartEbtSessionFor)
	if err != nil {
		return errors.Wrap(err, "error waiting for a session")
	}
	if !ok {
		return replication.ErrPeerDoesNotSupportEBT
	}
	return nil
}

func (r Replicator) replicateSelf(ctx context.Context, peer transport.Peer) {
	if err := r.selfCreateHistoryStreamReplicator.ReplicateSelf(ctx, peer); err != nil {
		r.logger.WithError(err).Error("error replicating self")
	}
}

func (r Replicator) HandleIncoming(ctx context.Context, version int, format messages.EbtReplicateFormat, stream Stream) error {
	if version != ebtReplicateVersion {
		return errors.New("invalid ebt version")
	}

	if format != ebtReplicateFormat {
		return errors.New("invalid ebt format")
	}

	connectionId, ok := rpc.GetConnectionIdFromContext(ctx)
	if !ok {
		return errors.New("connection id not found in context")
	}

	r.logger.WithField("connection_id", connectionId).Debug("incoming EBT session")

	done, err := r.tracker.OpenSession(connectionId)
	if err != nil {
		return errors.Wrap(err, "failed to mark local session as open")
	}
	defer done()

	return r.runner.HandleStream(ctx, stream)
}

func (r Replicator) openEbtStream(ctx context.Context, peer transport.Peer) (rpc.ResponseStream, error) {
	args, err := messages.NewEbtReplicateArguments(ebtReplicateVersion, ebtReplicateFormat)
	if err != nil {
		return nil, errors.Wrap(err, "error creating arguments")
	}

	req, err := messages.NewEbtReplicate(args)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the request")
	}

	rs, err := peer.Conn().PerformRequest(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "error performing the request")
	}

	return rs, nil
}
