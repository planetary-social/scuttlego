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

type PeerInitializer struct {
	handshaker            boxstream.Handshaker
	requestHandler        rpc.RequestHandler
	connectionIdGenerator *rpc.ConnectionIdGenerator
	logger                logging.Logger
}

func NewPeerInitializer(
	handshaker boxstream.Handshaker,
	requestHandler rpc.RequestHandler,
	connectionIdGenerator *rpc.ConnectionIdGenerator,
	logger logging.Logger,
) *PeerInitializer {
	return &PeerInitializer{
		handshaker:            handshaker,
		requestHandler:        requestHandler,
		connectionIdGenerator: connectionIdGenerator,
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

	raw := transport.NewRawConnection(boxStream, logger)

	rpcConn, err := rpc.NewConnection(ctx, connectionId, wasInitiatedByRemote, raw, i.requestHandler, logger)
	if err != nil {
		return Peer{}, errors.Wrap(err, "failed to establish an RPC connection")
	}

	peer, err := NewPeer(boxStream.Remote(), rpcConn)
	if err != nil {
		return Peer{}, errors.Wrap(err, "error creating a peer")
	}

	return peer, nil
}
