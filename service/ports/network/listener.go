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
	InitializeServerPeer(rwc io.ReadWriteCloser) (transport.Peer, error)
}

type Listener struct {
	initializer ServerPeerInitializer
	app         app.Application
	address     string
	logger      logging.Logger
}

func NewListener(initializer ServerPeerInitializer, app app.Application, address string, logger logging.Logger) (*Listener, error) {
	return &Listener{
		initializer: initializer,
		app:         app,
		address:     address,
		logger:      logger.New("listener"),
	}, nil
}

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
	p, err := l.initializer.InitializeServerPeer(conn)
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
