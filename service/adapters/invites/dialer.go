package invites

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type Dialer interface {
	DialWithInitializer(ctx context.Context, initializer network.ClientPeerInitializer, remote identity.Public, addr network.Address) (transport.Peer, error)
}

type CurrentTimeProvider interface {
	Get() time.Time
}

type InviteDialer struct {
	dialer                Dialer
	networkKey            boxstream.NetworkKey
	requestHandler        rpc.RequestHandler
	connectionIdGenerator *rpc.ConnectionIdGenerator
	currentTimeProvider   CurrentTimeProvider
	logger                logging.Logger
}

func NewInviteDialer(
	dialer Dialer,
	networkKey boxstream.NetworkKey,
	requestHandler rpc.RequestHandler,
	connectionIdGenerator *rpc.ConnectionIdGenerator,
	currentTimeProvider CurrentTimeProvider,
	logger logging.Logger,
) *InviteDialer {
	return &InviteDialer{
		dialer:                dialer,
		networkKey:            networkKey,
		requestHandler:        requestHandler,
		connectionIdGenerator: connectionIdGenerator,
		currentTimeProvider:   currentTimeProvider,
		logger:                logger,
	}
}

func (h *InviteDialer) Dial(ctx context.Context, local identity.Private, remote identity.Public, address network.Address) (transport.Peer, error) {
	handshaker, err := boxstream.NewHandshaker(local, h.networkKey, h.currentTimeProvider)
	if err != nil {
		return transport.Peer{}, errors.Wrap(err, "could not create a handshaker")
	}

	initializer := transport.NewPeerInitializer(handshaker, h.requestHandler, h.connectionIdGenerator, newNoopPeerHandler(), h.logger)

	peer, err := h.dialer.DialWithInitializer(ctx, initializer, remote, address)
	if err != nil {
		return transport.Peer{}, errors.Wrap(err, "failed to dial")
	}

	return peer, nil
}

type noopPeerHandler struct {
}

func newNoopPeerHandler() *noopPeerHandler {
	return &noopPeerHandler{}
}

func (n noopPeerHandler) HandleNewPeer(ctx context.Context, peer transport.Peer) {
}
