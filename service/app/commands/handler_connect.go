package commands

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
)

// Connect tries to initiate the connection to the specified node. This is most likely useful when you want to
// explicitly stimulate the program to talk to a specific node. Normally connections are initiated and managed
// automatically. Executing this command doesn't necessarily mean that a new connection will be established, for
// example the underlying implementation may decide not to do this if the connection with the specified identity
// already exists.
type Connect struct {
	// Remote is the identity of the remote node.
	Remote identity.Public

	// Address is the address of the remote node.
	Address network.Address
}

type ConnectHandler struct {
	peerManager PeerManager
	logger      logging.Logger
}

func NewConnectHandler(
	peerManager PeerManager,
	logger logging.Logger,
) *ConnectHandler {
	return &ConnectHandler{
		peerManager: peerManager,
		logger:      logger,
	}
}

func (h *ConnectHandler) Handle(ctx context.Context, cmd Connect) error {
	if err := h.peerManager.Connect(ctx, cmd.Remote, cmd.Address); err != nil {
		return errors.Wrap(err, "error initiating the connection")
	}

	return nil
}
