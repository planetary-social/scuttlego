// Package pubsub receives internal events.
package pubsub

import (
	"context"

	"github.com/planetary-social/scuttlego/service/adapters/pubsub"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux"
)

// RequestSubscriber receives internal events containing arriving RPC requests and passes
// them to the RPC mux.
//
// This is done because while it makes perfect sense for the RPC mux to be a
// port our program creates new RPC connections (which are sources of RPC
// requests) not only in the ports layer but also in the application layer (when
// establishing new connections). This is different from common programs which
// only receive incoming connections which e.g. provide REST requests. Since
// both the fact that connections must exist in the application layer and the
// fact that the RPC mux is a port make sense then we need a mechanism such as
// pub sub to bridge the two parts of the program together as the application
// layer can't directly drive ports.
type RequestSubscriber struct {
	pubsub *pubsub.RequestPubSub
	mux    *mux.Mux
}

func NewRequestSubscriber(pubsub *pubsub.RequestPubSub, mux *mux.Mux) *RequestSubscriber {
	return &RequestSubscriber{
		pubsub: pubsub,
		mux:    mux,
	}
}

// Run keeps receiving RPC requests from the pubsub and passing them to the RPC
// mux until the context is closed.
func (p *RequestSubscriber) Run(ctx context.Context) error {
	requests := p.pubsub.SubscribeToRequests(ctx)

	for request := range requests {
		p.mux.HandleRequest(request.Ctx, request.Rw, request.Req)
	}

	return nil
}
