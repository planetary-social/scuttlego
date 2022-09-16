package replication

import (
	"context"
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"time"
)

var ErrPeerDoesNotSupportEBT = errors.New("peer does not support ebt")

type EpidemicBroadcastTreesReplicator interface {
	// Replicate returns ErrPeerDoesNotSupportEBT if the peer doesn't support
	// EBT replication.
	Replicate(ctx context.Context, peer transport.Peer) error
}

type CreateHistoryStreamReplicator interface {
	Replicate(ctx context.Context, peer transport.Peer) error
}

type Negotiator struct {
	logger        logging.Logger
	ebtReplicator EpidemicBroadcastTreesReplicator
	chsReplicator CreateHistoryStreamReplicator
}

func NewNegotiator(logger logging.Logger) *Negotiator {
	return &Negotiator{logger: logger}
}

func (n Negotiator) Replicate(ctx context.Context, peer transport.Peer) error {
	for {
		if err := n.initiateReplication(ctx, peer); err != nil {
			n.peerLogger(peer).
				WithError(err).
				Debug("failed to initiate replication")
		}

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (n Negotiator) initiateReplication(ctx context.Context, peer transport.Peer) error {
	if err := n.ebtReplicator.Replicate(ctx, peer); err != nil {
		if errors.Is(err, ErrPeerDoesNotSupportEBT) {
			return n.fallbackToCreateHistoryStream(ctx, peer)
		}
		return errors.Wrap(err, "EBT replicator error")
	}
	return nil
}

func (n Negotiator) fallbackToCreateHistoryStream(ctx context.Context, peer transport.Peer) error {
	n.peerLogger(peer).Debug("peer does not support EBT replication, falling back to create history stream")
	if err := n.chsReplicator.Replicate(ctx, peer); err != nil {
		return errors.Wrap(err, "CHS replicator error")
	}
	return nil
}

func (n Negotiator) peerLogger(peer transport.Peer) logging.Logger {
	return n.logger.WithField("peer", peer)
}
