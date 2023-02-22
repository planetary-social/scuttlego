package transport

import (
	"context"
	"io"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
)

type NewPeerHandler interface {
	HandleNewPeer(ctx context.Context, peer Peer)
}

type PeerInitializer struct {
	handshaker            boxstream.Handshaker
	requestHandler        rpc.RequestHandler
	connectionIdGenerator *rpc.ConnectionIdGenerator
	newPeerHandler        NewPeerHandler
	logger                logging.Logger
}

func NewPeerInitializer(
	handshaker boxstream.Handshaker,
	requestHandler rpc.RequestHandler,
	connectionIdGenerator *rpc.ConnectionIdGenerator,
	newPeerHandler NewPeerHandler,
	logger logging.Logger,
) *PeerInitializer {
	return &PeerInitializer{
		handshaker:            handshaker,
		requestHandler:        requestHandler,
		connectionIdGenerator: connectionIdGenerator,
		newPeerHandler:        newPeerHandler,
		logger:                logger,
	}
}

func (i PeerInitializer) InitializeServerPeer(ctx context.Context, rwc io.ReadWriteCloser) (Peer, error) {
	boxStream, err := i.handshaker.OpenServerStream(rwc)
	if err != nil {
		return Peer{}, errors.Wrap(err, "failed to open a server stream")
	}

	return i.initializePeer(ctx, boxStream, true)
}

func (i PeerInitializer) InitializeClientPeer(ctx context.Context, rwc io.ReadWriteCloser, remote identity.Public) (Peer, error) {
	boxStream, err := i.handshaker.OpenClientStream(rwc, remote)
	if err != nil {
		return Peer{}, errors.Wrap(err, "failed to open a client stream")
	}

	return i.initializePeer(ctx, boxStream, false)
}

func (i PeerInitializer) initializePeer(ctx context.Context, boxStream *boxstream.Stream, wasInitiatedByRemote bool) (Peer, error) {
	connectionId := i.connectionIdGenerator.Generate()

	ctx = logging.AddToLoggingContext(ctx, logging.ConnectionIdContextLabel, connectionId)
	ctx = logging.AddToLoggingContext(ctx, logging.PeerIdContextLabel, boxStream.Remote().String())

	logger := i.logger.WithCtx(ctx)

	ctx = rpc.PutRemoteIdentityInContext(ctx, boxStream.Remote())
	ctx = rpc.PutConnectionIdInContext(ctx, connectionId)

	raw := transport.NewRawConnection(boxStream, logger)

	rpcConn, err := rpc.NewConnection(connectionId, wasInitiatedByRemote, raw, i.requestHandler, logger)
	if err != nil {
		return Peer{}, errors.Wrap(err, "failed to establish an RPC connection")
	}

	peer, err := NewPeer(boxStream.Remote(), rpcConn)
	if err != nil {
		return Peer{}, errors.Wrap(err, "error creating a peer")
	}

	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		i.newPeerHandler.HandleNewPeer(ctx, peer)

		if err := rpcConn.Loop(ctx); err != nil {
			logger.WithError(err).Debug("connection loop exited")
		}
	}()

	return peer, nil
}
