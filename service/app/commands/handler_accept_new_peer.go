package commands

import (
	"github.com/planetary-social/go-ssb/service/domain/transport"
)

type AcceptNewPeer struct {
	Peer transport.Peer
}

type AcceptNewPeerHandler struct {
	peerHandler NewPeerHandler
}

func NewAcceptNewPeerHandler(
	peerHandler NewPeerHandler,
) *AcceptNewPeerHandler {
	return &AcceptNewPeerHandler{
		peerHandler: peerHandler,
	}
}

func (h *AcceptNewPeerHandler) Handle(cmd AcceptNewPeer) error {
	h.peerHandler.HandleNewPeer(cmd.Peer)
	return nil
}
