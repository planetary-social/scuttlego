package commands

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type MessageReplicator interface {
	Replicate(ctx context.Context, peer transport.Peer) error
}

type BlobReplicator interface {
	Replicate(ctx context.Context, peer transport.Peer) error
}

type RoomScanner interface {
	Run(ctx context.Context, peer transport.Peer) error
}

type AcceptNewPeerHandler struct {
	peerManager       PeerManager
	messageReplicator MessageReplicator
	blobReplicator    BlobReplicator
	roomScanner       RoomScanner
	logger            logging.Logger
}

func NewAcceptNewPeerHandler(
	peerManager PeerManager,
	messageReplicator MessageReplicator,
	blobReplicator BlobReplicator,
	roomScanner RoomScanner,
	logger logging.Logger,
) *AcceptNewPeerHandler {
	return &AcceptNewPeerHandler{
		peerManager:       peerManager,
		messageReplicator: messageReplicator,
		blobReplicator:    blobReplicator,
		roomScanner:       roomScanner,
		logger:            logger,
	}
}

func (h *AcceptNewPeerHandler) Handle(ctx context.Context, peer transport.Peer) {
	h.logger.WithField("peer", peer).Debug("accepting a new peer")

	h.peerManager.TrackPeer(ctx, peer)
	go h.processConnection(ctx, peer)
}

func (h *AcceptNewPeerHandler) processConnection(ctx context.Context, peer transport.Peer) {
	if err := h.runTasks(ctx, peer); err != nil {
		h.logger.WithError(err).WithField("peer", peer).Debug("all tasks ended")
	}
}

func (h *AcceptNewPeerHandler) runTasks(ctx context.Context, peer transport.Peer) error {
	ch := make(chan error)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tasks := 0

	h.startTask(&tasks, ctx, peer, ch, h.messageReplicator.Replicate, "message replication")
	h.startTask(&tasks, ctx, peer, ch, h.blobReplicator.Replicate, "blob replication")
	h.startTask(&tasks, ctx, peer, ch, h.roomScanner.Run, "room scanner")

	var result error
	for i := 0; i < tasks; i++ {
		result = multierror.Append(result, <-ch)
	}
	return result
}

func (h *AcceptNewPeerHandler) startTask(
	tasks *int,
	ctx context.Context,
	peer transport.Peer,
	ch chan<- error,
	fn func(ctx context.Context, peer transport.Peer) error,
	taskName string,
) {
	peerLogger := h.logger.WithField("peer", peer)
	*tasks = *tasks + 1
	go func() {
		err := fn(ctx, peer)
		peerLogger.WithError(err).WithField("task", taskName).Debug("task terminating")
		ch <- err
	}()
}
