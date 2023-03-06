package mocks

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type PeerInitializerMock struct {
	InitializeServerPeerReturnValue transport.Peer
	initializeServerPeerCalls       atomic.Int32
}

func NewPeerInitializerMock() *PeerInitializerMock {
	return &PeerInitializerMock{}
}

func (p *PeerInitializerMock) InitializeServerPeer(ctx context.Context, rwc io.ReadWriteCloser) (transport.Peer, error) {
	p.initializeServerPeerCalls.Add(1)
	return p.InitializeServerPeerReturnValue, nil
}

func (p *PeerInitializerMock) InitializeServerPeerCalls() int {
	return int(p.initializeServerPeerCalls.Load())
}
