package tunnel

import (
	"context"
	"io"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type ClientPeerInitializer interface {
	// InitializeClientPeer initializes outgoing connections by performing a
	// handshake and establishing an RPC connection using the provided
	// ReadWriteCloser. Context is used as the RPC connection context.
	InitializeClientPeer(ctx context.Context, rwc io.ReadWriteCloser, remote identity.Public) (transport.Peer, error)
}

type Dialer struct {
	initializer ClientPeerInitializer
}

func NewDialer(initializer ClientPeerInitializer) *Dialer {
	return &Dialer{initializer: initializer}
}

func (d *Dialer) DialViaRoom(ctx context.Context, portal transport.Peer, target identity.Public) (transport.Peer, error) {
	// todo timeout?
	portalRef, err := refs.NewIdentityFromPublic(portal.Identity())
	if err != nil {
		return transport.Peer{}, errors.Wrap(err, "error creating portal identity ref")
	}

	targetRef, err := refs.NewIdentityFromPublic(target)
	if err != nil {
		return transport.Peer{}, errors.Wrap(err, "error creating target identity ref")
	}

	arguments, err := messages.NewTunnelConnectToPortalArguments(portalRef, targetRef)
	if err != nil {
		return transport.Peer{}, errors.Wrap(err, "error creating arguments")
	}

	request, err := messages.NewTunnelConnectToPortal(arguments)
	if err != nil {
		return transport.Peer{}, errors.Wrap(err, "error creating a request")
	}

	ctx, cancel := context.WithCancel(ctx)

	stream, err := portal.Conn().PerformRequest(ctx, request)
	if err != nil {
		cancel()
		return transport.Peer{}, errors.Wrap(err, "error performing the request")
	}

	rwc := NewResponseStreamReadWriteCloserAdapter(stream, cancel)
	peer, err := d.initializer.InitializeClientPeer(ctx, rwc, target)
	if err != nil {
		cancel()
		return transport.Peer{}, errors.Wrap(err, "error performing the request")
	}

	return peer, nil
}
