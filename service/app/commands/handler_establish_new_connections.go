package commands

import (
	"github.com/boreq/errors"
)

type EstablishNewConnectionsHandler struct {
	peerManager PeerManager
}

func NewEstablishNewConnectionsHandler(
	peerManager PeerManager,
) *EstablishNewConnectionsHandler {
	return &EstablishNewConnectionsHandler{
		peerManager: peerManager,
	}
}

func (h *EstablishNewConnectionsHandler) Handle() error {
	if err := h.peerManager.EstablishNewConnections(); err != nil {
		return errors.Wrap(err, "error calling peer manager")
	}

	return nil
}
