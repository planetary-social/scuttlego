package pubsub

import (
	"context"

	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
)

type Request struct {
	Ctx context.Context
	Req *rpc.Request
	Rw  *rpc.ResponseWriter
}

type RequestPubSub struct {
	pubsub *GoChannelPubSub[Request]
}

func NewRequestPubSub() *RequestPubSub {
	return &RequestPubSub{
		pubsub: NewGoChannelPubSub[Request](),
	}
}

func (m *RequestPubSub) HandleRequest(ctx context.Context, req *rpc.Request, rw *rpc.ResponseWriter) {
	m.pubsub.Publish(Request{Ctx: ctx, Req: req, Rw: rw})
}

func (m *RequestPubSub) SubscribeToRequests(ctx context.Context) <-chan Request {
	return m.pubsub.Subscribe(ctx)
}
