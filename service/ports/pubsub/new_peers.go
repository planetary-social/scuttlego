// Package pubsub receives internal events.
package pubsub

import (
	"context"

	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/adapters/pubsub"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type AcceptNewPeerCommandHandler interface {
	Handle(ctx context.Context, peer transport.Peer)
}

type NewPeerSubscriber struct {
	pubsub  *pubsub.NewPeerPubSub
	handler AcceptNewPeerCommandHandler
	logger  logging.Logger
}

func NewNewPeerSubscriber(
	pubsub *pubsub.NewPeerPubSub,
	handler AcceptNewPeerCommandHandler,
	logger logging.Logger,
) *NewPeerSubscriber {
	return &NewPeerSubscriber{
		pubsub:  pubsub,
		handler: handler,
		logger:  logger,
	}
}

func (p *NewPeerSubscriber) Run(ctx context.Context) error {
	for newPeer := range p.pubsub.SubscribeToRequests(ctx) {
		p.handler.Handle(newPeer.Ctx, newPeer.Peer)
	}

	return nil
}
