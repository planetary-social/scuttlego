// Package network handles incoming network connections.
package network

import (
	"context"
	"io"
	"net"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/app"
	"github.com/planetary-social/go-ssb/service/app/commands"
	"github.com/planetary-social/go-ssb/service/domain/transport"
)

type ServerPeerInitializer interface {
	// InitializeServerPeer initializes incoming connections by performing a
	// handshake and establishing an RPC connection using the provided
	// ReadWriteCloser. Context is used as the RPC connection context.
	InitializeServerPeer(ctx context.Context, rwc io.ReadWriteCloser) (transport.Peer, error)
}

// Listener handles incoming TCP connections initiated by other peers,
// initializes them and passes them to the AcceptNewPeer command.
type Listener struct {
	initializer   ServerPeerInitializer
	app           app.Application
	address       string
	logger        logging.Logger
	connectionCtx context.Context
}

// NewListener creates a new listener which listens on the provided address. The
// address should be formatted in the way which can be handled by the net
// package e.g. ":8008". The provided context is used to initiate the peer
// connections.
func NewListener(ctx context.Context, initializer ServerPeerInitializer, app app.Application, address string, logger logging.Logger) (*Listener, error) {
	return &Listener{
		initializer:   initializer,
		app:           app,
		address:       address,
		logger:        logger.New("listener"),
		connectionCtx: ctx,
	}, nil
}

// ListenAndServe starts listening and keeps accepting connections and
// processing them until the context is closed. The context passed to this
// function is not used to initiate RPC connections.
func (l *Listener) ListenAndServe(ctx context.Context) error {
	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", l.address)
	if err != nil {
		return errors.Wrap(err, "could not start a listener")
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return errors.Wrap(err, "could not accept a connection")
		}

		go l.handleNewConnection(conn)
	}
}

func (l *Listener) handleNewConnection(conn net.Conn) {
	p, err := l.initializer.InitializeServerPeer(l.connectionCtx, conn)
	if err != nil {
		conn.Close()
		l.logger.WithError(err).Debug("could not init a peer")
		return
	}

	err = l.app.Commands.AcceptNewPeer.Handle(commands.AcceptNewPeer{Peer: p})
	if err != nil {
		conn.Close()
		l.logger.WithError(err).Debug("could not accept a new peer")
		return
	}
}
