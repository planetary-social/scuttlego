package pubsub

import (
	"context"

	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type Request struct {
	Ctx context.Context
	Rw  rpc.ResponseWriter
	Req *rpc.Request
}

type RequestPubSub struct {
	pubsub *GoChannelPubSub[Request]
}

func NewRequestPubSub() *RequestPubSub {
	return &RequestPubSub{
		pubsub: NewGoChannelPubSub[Request](),
	}
}

func (m *RequestPubSub) HandleRequest(ctx context.Context, rw rpc.ResponseWriter, req *rpc.Request) {
	m.pubsub.Publish(
		Request{
			Ctx: ctx,
			Rw:  rw,
			Req: req,
		},
	)
}

func (m *RequestPubSub) SubscribeToRequests(ctx context.Context) <-chan Request {
	return m.pubsub.Subscribe(ctx)
}
