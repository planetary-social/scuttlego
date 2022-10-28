package commands

import (
	"context"
	"io"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

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

type AcceptTunnelConnectHandler struct {
}

func NewAcceptTunnelConnectHandler() *AcceptTunnelConnectHandler {
	return &AcceptTunnelConnectHandler{}
}

func (h *AcceptTunnelConnectHandler) Handle(ctx context.Context, cmd AcceptTunnelConnect) error {
	return errors.New("not implemented")
}
