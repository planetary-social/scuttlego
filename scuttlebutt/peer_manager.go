package scuttlebutt

import (
	"context"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/network"
	"github.com/planetary-social/go-ssb/network/rpc"
)

type RPC interface {
	Serve(ctx context.Context, conn *rpc.Connection) error
}

type Replicator interface {
	Replicate(ctx context.Context, peer network.Peer) error
}

type PeerManager struct {
	rpc        RPC
	replicator Replicator
	logger     logging.Logger
}

func NewPeerManager(rpc RPC, replicator Replicator, logger logging.Logger) *PeerManager {
	return &PeerManager{
		rpc:        rpc,
		replicator: replicator,
		logger:     logger.New("peer_manager"),
	}
}

func (p PeerManager) HandleNewPeer(peer network.Peer) error {
	p.logger.Debug("handling a new peer")

	go p.processConnection(peer)
	return nil
}

func (p PeerManager) processConnection(peer network.Peer) {
	if err := p.handleConnection(peer); err != nil {
		p.logger.WithError(err).WithField("peer", peer).Debug("connection ended")
	}
}

func (p PeerManager) handleConnection(peer network.Peer) error {
	ch := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())

	tasks := 0

	tasks++
	go func() {
		defer cancel()
		defer p.logger.Debug("serve task terminating")
		if err := p.rpc.Serve(ctx, peer.Conn()); err != nil {
			ch <- err
		}
	}()

	tasks++
	go func() {
		defer cancel()
		defer p.logger.Debug("replicate task terminating")
		if err := p.replicator.Replicate(ctx, peer); err != nil {
			ch <- err
		}
	}()

	var result error
	for i := 0; i < tasks; i++ {
		result = multierror.Append(result, <-ch)
	}
	return result
}
