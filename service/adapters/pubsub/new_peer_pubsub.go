package pubsub

import (
	"context"

	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type NewPeer struct {
	Ctx  context.Context
	Peer transport.Peer
}

type NewPeerPubSub struct {
	pubsub *GoChannelPubSub[NewPeer]
}

func NewNewPeerPubSub() *NewPeerPubSub {
	return &NewPeerPubSub{
		pubsub: NewGoChannelPubSub[NewPeer](),
	}
}

func (m *NewPeerPubSub) HandleNewPeer(ctx context.Context, peer transport.Peer) {
	m.pubsub.Publish(
		NewPeer{
			Ctx:  ctx,
			Peer: peer,
		},
	)
}

func (m *NewPeerPubSub) SubscribeToRequests(ctx context.Context) <-chan NewPeer {
	return m.pubsub.Subscribe(ctx)
}
