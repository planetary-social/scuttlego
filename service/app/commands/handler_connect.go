package commands

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/network"
)

type Connect struct {
	Remote  identity.Public
	Address network.Address
}

type ConnectHandler struct {
	dialer      Dialer
	peerHandler NewPeerHandler
	logger      logging.Logger
}

func NewConnectHandler(
	dialer Dialer,
	peerHandler NewPeerHandler,
	logger logging.Logger,
) *ConnectHandler {
	return &ConnectHandler{
		dialer:      dialer,
		peerHandler: peerHandler,
		logger:      logger,
	}
}

func (h *ConnectHandler) Handle(cmd Connect) error {
	peer, err := h.dialer.Dial(cmd.Remote, cmd.Address)
	if err != nil {
		return errors.Wrap(err, "dial failed")
	}

	h.peerHandler.HandleNewPeer(peer)

	return nil
}
