package network

import (
	"io"

	"github.com/planetary-social/go-ssb/service/domain/identity"
	boxstream2 "github.com/planetary-social/go-ssb/service/domain/network/boxstream"
	rpc2 "github.com/planetary-social/go-ssb/service/domain/network/rpc"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
)

type PeerInitializer struct {
	handshaker     boxstream2.Handshaker
	requestHandler rpc2.RequestHandler
	logger         logging.Logger
}

func NewPeerInitializer(
	handshaker boxstream2.Handshaker,
	requestHandler rpc2.RequestHandler,
	logger logging.Logger,
) *PeerInitializer {
	return &PeerInitializer{
		handshaker:     handshaker,
		requestHandler: requestHandler,
		logger:         logger,
	}
}

func (i PeerInitializer) InitializeServerPeer(rwc io.ReadWriteCloser) (Peer, error) {
	boxStream, err := i.handshaker.OpenServerStream(rwc)
	if err != nil {
		return Peer{}, errors.Wrap(err, "failed to open a server stream")
	}

	return i.initializePeer(boxStream)
}

func (i PeerInitializer) InitializeClientPeer(rwc io.ReadWriteCloser, remote identity.Public) (Peer, error) {
	boxStream, err := i.handshaker.OpenClientStream(rwc, remote)
	if err != nil {
		return Peer{}, errors.Wrap(err, "failed to open a client stream")
	}

	return i.initializePeer(boxStream)
}

func (i PeerInitializer) initializePeer(boxStream *boxstream2.Stream) (Peer, error) {
	rpcConn, err := rpc2.NewConnection(boxStream, i.requestHandler, i.logger)
	if err != nil {
		return Peer{}, errors.Wrap(err, "failed to establish an RPC connection")
	}

	return NewPeer(boxStream.Remote(), rpcConn), nil
}
