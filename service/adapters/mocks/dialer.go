package mocks

import (
	"context"
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type DialerMock struct {
	peers map[string]transport.Peer
}

func NewDialerMock() *DialerMock {
	return &DialerMock{
		peers: make(map[string]transport.Peer),
	}
}

func (d *DialerMock) DialWithInitializer(ctx context.Context, initializer network.ClientPeerInitializer, remote identity.Public, addr network.Address) (transport.Peer, error) {
	return transport.Peer{}, errors.New("not implemented")
}

func (d *DialerMock) Dial(ctx context.Context, remote identity.Public, address network.Address) (transport.Peer, error) {
	p, ok := d.peers[d.key(remote, address)]
	if !ok {
		return transport.Peer{}, fmt.Errorf("peer for identity '%s' and address '%s' was not mocked", remote, address)
	}
	return p, nil
}

func (d *DialerMock) MockPeer(remote identity.Public, address network.Address, connection transport.Connection) {
	d.peers[d.key(remote, address)] = transport.NewPeer(
		remote,
		connection,
	)

}

func (d *DialerMock) key(remote identity.Public, address network.Address) string {
	return fmt.Sprintf("%s-%s", remote.String(), address.String())

}
