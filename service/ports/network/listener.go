// Package network handles incoming network connections.
package network

import (
	"context"
	"io"
	"net"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/transport"
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
	initializer ServerPeerInitializer
	address     string
	logger      logging.Logger
}

// NewListener creates a new listener which listens on the provided address. The
// address should be formatted in the way which can be handled by the net
// package e.g. ":8008".
func NewListener(
	initializer ServerPeerInitializer,
	address string,
	logger logging.Logger,
) (*Listener, error) {
	return &Listener{
		initializer: initializer,
		address:     address,
		logger:      logger.New("listener"),
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

	go func() {
		<-ctx.Done()
		if err := listener.Close(); err != nil {
			l.logger.WithError(err).Error("error closing the listener")
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return errors.Wrap(err, "could not accept a connection")
		}

		go l.handleNewConnection(ctx, conn)
	}
}

func (l *Listener) handleNewConnection(ctx context.Context, conn net.Conn) {
	_, err := l.initializer.InitializeServerPeer(ctx, conn)
	if err != nil {
		conn.Close()
		l.logger.WithError(err).Debug("could not init a peer")
		return
	}
}
