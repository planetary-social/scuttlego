package commands

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type AcceptNewPeer struct {
	peer transport.Peer
}

func NewAcceptNewPeer(peer transport.Peer) (AcceptNewPeer, error) {
	if peer.IsZero() {
		return AcceptNewPeer{}, errors.New("zero value of peer")
	}
	return AcceptNewPeer{peer: peer}, nil
}

func (a AcceptNewPeer) Peer() transport.Peer {
	return a.peer
}

func (a AcceptNewPeer) IsZero() bool {
	return a.peer.IsZero()
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
	if cmd.IsZero() {
		return errors.New("zero value of cmd")
	}

	h.peerHandler.HandleNewPeer(cmd.Peer())
	return nil
}
