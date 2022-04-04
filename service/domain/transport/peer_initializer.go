package transport

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/transport/boxstream"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc/transport"
)

type PeerInitializer struct {
	handshaker     boxstream.Handshaker
	requestHandler rpc.RequestHandler
	logger         logging.Logger
}

func NewPeerInitializer(
	handshaker boxstream.Handshaker,
	requestHandler rpc.RequestHandler,
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

func (i PeerInitializer) initializePeer(boxStream *boxstream.Stream) (Peer, error) {
	logger, err := i.peerLogger(boxStream)
	if err != nil {
		return Peer{}, errors.Wrap(err, "failed to create a peer logger")
	}

	raw := transport.NewRawConnection(boxStream, logger)

	rpcConn, err := rpc.NewConnection(raw, i.requestHandler, logger)
	if err != nil {
		return Peer{}, errors.Wrap(err, "failed to establish an RPC connection")
	}

	return NewPeer(boxStream.Remote(), rpcConn), nil
}

func (i PeerInitializer) peerLogger(boxStream *boxstream.Stream) (logging.Logger, error) {
	ref, err := refs.NewIdentityFromPublic(boxStream.Remote())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create an identity ref")
	}

	correlationId := fmt.Sprintf("peer-%d", rand.Int())

	return i.logger.New(correlationId).WithField("id", ref.String()), nil
}
