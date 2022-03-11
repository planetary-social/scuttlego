package pubsub

import (
	"context"

	"github.com/planetary-social/go-ssb/service/adapters/pubsub"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
)

type PubSub struct {
	pubsub *pubsub.RequestPubSub
	mux    *rpc.Mux
}

func NewPubSub(pubsub *pubsub.RequestPubSub, mux *rpc.Mux) *PubSub {
	return &PubSub{
		pubsub: pubsub,
		mux:    mux,
	}
}

func (p *PubSub) Run(ctx context.Context) error {
	requests := p.pubsub.SubscribeToRequests(ctx)

	for request := range requests {
		go p.mux.HandleRequest(request.Req, request.Rw)
	}

	return nil
}
