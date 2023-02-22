package mocks

import (
	"context"
	"sync/atomic"

	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type NewPeerHandlerMock struct {
	handleNewPeerCalls atomic.Int32
}

func NewNewPeerHandlerMock() *NewPeerHandlerMock {
	return &NewPeerHandlerMock{}
}

func (n *NewPeerHandlerMock) HandleNewPeer(ctx context.Context, peer transport.Peer) {
	n.handleNewPeerCalls.Add(1)
}

func (n *NewPeerHandlerMock) HandleNewPeerCalls() int {
	return int(n.handleNewPeerCalls.Load())
}
