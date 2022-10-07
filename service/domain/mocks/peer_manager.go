package mocks

import (
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type PeerManagerMock struct {
	connectViaRoomCalls []PeerManagerConnectViaRoomCall
	peersReturnValue    []transport.Peer
}

func NewPeerManagerMock() *PeerManagerMock {
	return &PeerManagerMock{}
}

func (p *PeerManagerMock) Connect(remote identity.Public, address network.Address) error {
	//TODO implement me
	panic("implement me")
}

func (p *PeerManagerMock) ConnectViaRoom(portal transport.Peer, target identity.Public) error {
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

func (p *PeerManagerMock) EstablishNewConnections() error {
	//TODO implement me
	panic("implement me")
}

func (p *PeerManagerMock) ProcessNewLocalDiscovery(remote identity.Public, address network.Address) error {
	//TODO implement me
	panic("implement me")
}

func (p PeerManagerMock) Peers() []transport.Peer {
	return p.peersReturnValue
}

func (p PeerManagerMock) MockPeers(peers []transport.Peer) {
	p.peersReturnValue = peers
}

type PeerManagerConnectViaRoomCall struct {
	Portal transport.Peer
	Target identity.Public
}
