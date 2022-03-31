package network

import (
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
	logger      logging.Logger
}

func NewListener(initializer ServerPeerInitializer, app app.Application, logger logging.Logger) (*Listener, error) {
	return &Listener{
		initializer: initializer,
		app:         app,
		logger:      logger.New("listener"),
	}, nil
}

func (l *Listener) ListenAndServe() error {
	listener, err := net.Listen("tcp", ":8001") // todo configure the address
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
		p.Conn().Close()
		l.logger.WithError(err).Debug("could not init a peer")
		return
	}

	err = l.app.Commands.AcceptNewPeer.Handle(commands.AcceptNewPeer{Peer: p})
	if err != nil {
		p.Conn().Close()
		l.logger.WithError(err).Debug("could not accept a new peer")
		return
	}
}
