package network

import (
	"io"
	"net"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
)

type ServerPeerInitializer interface {
	InitializeServerPeer(rwc io.ReadWriteCloser) (Peer, error)
}

type NewPeerHandler interface {
	HandleNewPeer(Peer) error
}

type Listener struct {
	initializer ServerPeerInitializer
	peerHandler NewPeerHandler
	logger      logging.Logger
}

func NewListener(initializer ServerPeerInitializer, peerHandler NewPeerHandler, logger logging.Logger) (Listener, error) {
	return Listener{
		initializer: initializer,
		peerHandler: peerHandler,
		logger:      logger.New("listener"),
	}, nil
}

func (l Listener) ListenAndServe() error {
	l.logger.Debug("starting listening")

	listener, err := net.Listen("tcp", ":8001")
	if err != nil {
		return errors.Wrap(err, "could not start a listener")
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return errors.Wrap(err, "could not accept a connection")
		}

		go func() {
			p, err := l.initializer.InitializeServerPeer(conn)
			if err != nil {
				p.Conn().Close()
				l.logger.WithError(err).Debug("could not init a peer")
			}

			if err := l.peerHandler.HandleNewPeer(p); err != nil {
				p.Conn().Close()
				l.logger.WithError(err).Debug("could not handle a peer")
			}
		}()
	}
}
