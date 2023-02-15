package commands

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
)

// Connect tries to initiate the connection to the specified node. This is most
// likely useful when you want to explicitly stimulate the program to talk to a
// specific node. Normally connections are initiated and managed automatically.
// Executing this command doesn't necessarily mean that a new connection will be
// established, for example the underlying implementation may decide not to do
// this if the connection with the specified identity already exists.
type Connect struct {
	// remote is the identity of the remote node.
	remote identity.Public

	// address is the address of the remote node.
	address network.Address
}

func NewConnect(remote identity.Public, address network.Address) (Connect, error) {
	if remote.IsZero() {
		return Connect{}, errors.New("zero value of remote")
	}
	if address.IsZero() {
		return Connect{}, errors.New("zero value of address")
	}
	return Connect{remote: remote, address: address}, nil
}

func (c Connect) Remote() identity.Public {
	return c.remote
}

func (c Connect) Address() network.Address {
	return c.address
}

func (c Connect) IsZero() bool {
	return c.remote.IsZero()
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

func (h *ConnectHandler) Handle(cmd Connect) error {
	if cmd.IsZero() {
		return errors.New("zero value of cmd")
	}

	if err := h.peerManager.Connect(cmd.Remote(), cmd.Address()); err != nil {
		return errors.Wrap(err, "connect error")
	}
	return nil
}
