package ebt

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

const (
	ebtReplicateVersion = 3
)

type Replicator struct {
	tracker *SessionTracker
	runner  *SessionRunner
}

func (r Replicator) Replicate(ctx context.Context, peer transport.Peer) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if !peer.Conn().WasInitiatedByRemote() {
		return r.startLocalSession(ctx, peer)
	}

	return r.tracker.WaitForSession(ctx, peer.Conn().Id())
}

func (r Replicator) startLocalSession(ctx context.Context, peer transport.Peer) error {
	done, err := r.tracker.OpenLocalSession(peer.Conn().Id())
	if err != nil {
		return errors.Wrap(err, "failed to mark local session as open")
	}
	defer done()

	rs, err := r.openEbtStream(ctx, peer)
	if err != nil {
		return errors.Wrap(err, "error starting the ebt session")
	}

	return r.runner.HandleStream(ctx, NewResponseStreamAdapter(rs))
}

func (r Replicator) openEbtStream(ctx context.Context, peer transport.Peer) (*rpc.ResponseStream, error) {
	args, err := messages.NewEbtReplicateArguments(ebtReplicateVersion, messages.EbtReplicateFormatClassic)
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
