package commands

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
)

type ProcessNewLocalDiscovery struct {
	remote  identity.Public
	address network.Address
}

func NewProcessNewLocalDiscovery(
	remote identity.Public,
	address network.Address,
) (ProcessNewLocalDiscovery, error) {
	if remote.IsZero() {
		return ProcessNewLocalDiscovery{}, errors.New("zero value of remote")
	}
	if address.IsZero() {
		return ProcessNewLocalDiscovery{}, errors.New("zero value of address")
	}
	return ProcessNewLocalDiscovery{remote: remote, address: address}, nil
}

func (p ProcessNewLocalDiscovery) Remote() identity.Public {
	return p.remote
}

func (p ProcessNewLocalDiscovery) Address() network.Address {
	return p.address
}

func (p ProcessNewLocalDiscovery) IsZero() bool {
	return p.remote.IsZero()
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
	if cmd.IsZero() {
		return errors.New("zero value of cmd")
	}

	if err := h.peerManager.ProcessNewLocalDiscovery(cmd.Remote(), cmd.Address()); err != nil {
		return errors.Wrap(err, "error calling peer manager")
	}
	return nil
}
