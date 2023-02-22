package commands

import (
	"context"
	"io"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type ServerPeerInitializer interface {
	InitializeServerPeer(ctx context.Context, rwc io.ReadWriteCloser) (transport.Peer, error)
}

type AcceptTunnelConnect struct {
	origin refs.Identity
	target refs.Identity
	portal refs.Identity
	rwc    io.ReadWriteCloser
}

func NewAcceptTunnelConnect(origin refs.Identity, target refs.Identity, portal refs.Identity, rwc io.ReadWriteCloser) (AcceptTunnelConnect, error) {
	if origin.IsZero() {
		return AcceptTunnelConnect{}, errors.New("zero value of origin")
	}
	if target.IsZero() {
		return AcceptTunnelConnect{}, errors.New("zero value of target")
	}
	if portal.IsZero() {
		return AcceptTunnelConnect{}, errors.New("zero value of portal")
	}
	if rwc == nil {
		return AcceptTunnelConnect{}, errors.New("rwc is nil")
	}
	return AcceptTunnelConnect{
		origin: origin,
		target: target,
		portal: portal,
		rwc:    rwc,
	}, nil
}

func (a AcceptTunnelConnect) Origin() refs.Identity {
	return a.origin
}

func (a AcceptTunnelConnect) Target() refs.Identity {
	return a.target
}

func (a AcceptTunnelConnect) Portal() refs.Identity {
	return a.portal
}

func (a AcceptTunnelConnect) Rwc() io.ReadWriteCloser {
	return a.rwc
}

func (a AcceptTunnelConnect) IsZero() bool {
	return a.origin.IsZero()
}

type AcceptTunnelConnectHandler struct {
	local       identity.Public
	initializer ServerPeerInitializer
}

func NewAcceptTunnelConnectHandler(
	local identity.Public,
	initializer ServerPeerInitializer,
) *AcceptTunnelConnectHandler {
	return &AcceptTunnelConnectHandler{
		local:       local,
		initializer: initializer,
	}
}

func (h *AcceptTunnelConnectHandler) Handle(ctx context.Context, cmd AcceptTunnelConnect) error {
	if cmd.IsZero() {
		return errors.New("zero value of cmd")
	}

	if !cmd.Target().Identity().Equal(h.local) {
		return errors.New("target doesn't match local identity")
	}

	_, err := h.initializer.InitializeServerPeer(ctx, cmd.Rwc())
	if err != nil {
		return errors.Wrap(err, "failed to initialize the peer")
	}

	return nil
}
