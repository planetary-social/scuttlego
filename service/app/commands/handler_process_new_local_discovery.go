package commands

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
)

type ProcessNewLocalDiscovery struct {
	Remote  identity.Public
	Address network.Address
}

type ProcessNewLocalDiscoveryHandler struct {
	peerManager PeerManager
}

func NewProcessNewLocalDiscoveryHandler(peerManager PeerManager) *ProcessNewLocalDiscoveryHandler {
	return &ProcessNewLocalDiscoveryHandler{
		peerManager: peerManager,
	}
}

func (h *ProcessNewLocalDiscoveryHandler) Handle(cmd ProcessNewLocalDiscovery) error {
	if err := h.peerManager.ProcessNewLocalDiscovery(cmd.Remote, cmd.Address); err != nil {
		return errors.Wrap(err, "error calling peer manager")
	}

	return nil
}
