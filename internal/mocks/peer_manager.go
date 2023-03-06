package mocks

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type PeerManagerMock struct {
	connectViaRoomCalls []PeerManagerConnectViaRoomCall
	peersReturnValue    []transport.Peer
	disconnectAllCalls  int
}

func NewPeerManagerMock() *PeerManagerMock {
	return &PeerManagerMock{}
}

func (p *PeerManagerMock) DisconnectAllCalls() int {
	return p.disconnectAllCalls
}

func (p *PeerManagerMock) DisconnectAll() error {
	p.disconnectAllCalls++
	return nil
}

func (p *PeerManagerMock) Connect(ctx context.Context, remote identity.Public, address network.Address) error {
	return errors.New("not implemented")
}

func (p *PeerManagerMock) ConnectViaRoom(ctx context.Context, portal transport.Peer, target identity.Public) error {
	p.connectViaRoomCalls = append(
		p.connectViaRoomCalls,
		PeerManagerConnectViaRoomCall{
			Portal: portal,
			Target: target,
		},
	)
	return nil
}

func (p *PeerManagerMock) ConnectViaRoomCalls() []PeerManagerConnectViaRoomCall {
	return p.connectViaRoomCalls
}

func (p *PeerManagerMock) EstablishNewConnections(ctx context.Context) error {
	return errors.New("not implemented")
}

func (p *PeerManagerMock) ProcessNewLocalDiscovery(ctx context.Context, remote identity.Public, address network.Address) error {
	return errors.New("not implemented")
}

func (p *PeerManagerMock) Peers() []transport.Peer {
	return p.peersReturnValue
}

func (p *PeerManagerMock) MockPeers(peers []transport.Peer) {
	p.peersReturnValue = peers
}

func (p *PeerManagerMock) TrackPeer(ctx context.Context, peer transport.Peer) {
	//TODO implement me
	panic("implement me")
}

type PeerManagerConnectViaRoomCall struct {
	Portal transport.Peer
	Target identity.Public
}
