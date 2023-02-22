package commands

import (
	"context"

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

func (h *EstablishNewConnectionsHandler) Handle(ctx context.Context) error {
	if err := h.peerManager.EstablishNewConnections(ctx); err != nil {
		return errors.Wrap(err, "error calling peer manager")
	}

	return nil
}
