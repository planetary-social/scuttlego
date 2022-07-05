package mocks

import (
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type PeerManagerMock struct {
	PeersReturnValue []transport.Peer
}

func NewPeerManagerMock() *PeerManagerMock {
	return &PeerManagerMock{}
}

func (p PeerManagerMock) Peers() []transport.Peer {
	return p.PeersReturnValue
}
