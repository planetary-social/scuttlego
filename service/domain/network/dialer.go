package network

import (
	"io"
	"net"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/identity"
)

type ClientPeerInitializer interface {
	InitializeClientPeer(rwc io.ReadWriteCloser, remote identity.Public) (Peer, error)
}

type Dialer struct {
	initializer ClientPeerInitializer
	logger      logging.Logger
}

func NewDialer(initializer ClientPeerInitializer, logger logging.Logger) (*Dialer, error) {
	return &Dialer{
		initializer: initializer,
		logger:      logger.New("network"),
	}, nil
}

func (d Dialer) DialWithInitializer(initializer ClientPeerInitializer, remote identity.Public, addr Address) (Peer, error) {
	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		return Peer{}, errors.Wrap(err, "could not dial")
	}

	peer, err := initializer.InitializeClientPeer(conn, remote)
	if err != nil {
		return Peer{}, errors.Wrap(err, "could not initialize a client peer")
	}

	return peer, nil
}

func (d Dialer) Dial(remote identity.Public, address Address) (Peer, error) {
	return d.DialWithInitializer(d.initializer, remote, address)
}
