package network

import (
	"io"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/identity"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/network/boxstream"
	"github.com/planetary-social/go-ssb/network/rpc"
)

type PeerInitializer struct {
	handshaker boxstream.Handshaker
	logger     logging.Logger
}

func NewPeerInitializer(handshaker boxstream.Handshaker, logger logging.Logger) *PeerInitializer {
	return &PeerInitializer{
		handshaker: handshaker,
		logger:     logger,
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

func (i PeerInitializer) initializePeer(boxStream *boxstream.Stream) (Peer, error) {
	rpcConn, err := rpc.NewConnection(boxStream, i.logger)
	if err != nil {
		return Peer{}, errors.Wrap(err, "failed to establish an RPC connection")
	}

	return NewPeer(boxStream.Remote(), rpcConn), nil
}
