package replication

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

var ErrPeerDoesNotSupportEBT = errors.New("peer does not support ebt")

type EpidemicBroadcastTreesReplicator interface {
	// Replicate should keep attempting to perform replication as long as the
	// context isn't closed. Returning an error implies that replication should
	// not restart. Returns ErrPeerDoesNotSupportEBT if the peer doesn't support
	// EBT replication.
	Replicate(ctx context.Context, peer transport.Peer) error
}

type CreateHistoryStreamReplicator interface {
	// Replicate should keep attempting to perform replication as long as the
	// context isn't closed. Returning an error implies that replication should
	// not restart.
	Replicate(ctx context.Context, peer transport.Peer) error
}

type Negotiator struct {
	logger        logging.Logger
	ebtReplicator EpidemicBroadcastTreesReplicator
	chsReplicator CreateHistoryStreamReplicator
}

func NewNegotiator(
	logger logging.Logger,
	ebtReplicator EpidemicBroadcastTreesReplicator,
	chsReplicator CreateHistoryStreamReplicator,
) *Negotiator {
	return &Negotiator{
		logger:        logger,
		ebtReplicator: ebtReplicator,
		chsReplicator: chsReplicator,
	}
}

func (n Negotiator) Replicate(ctx context.Context, peer transport.Peer) error {
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
