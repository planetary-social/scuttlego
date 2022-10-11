package pubsub

import (
	"context"

	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type Request struct {
	Ctx    context.Context
	Stream rpc.Stream
	Req    *rpc.Request
}

type RequestPubSub struct {
	pubsub *GoChannelPubSub[Request]
}

func NewRequestPubSub() *RequestPubSub {
	return &RequestPubSub{
		pubsub: NewGoChannelPubSub[Request](),
	}
}

func (m *RequestPubSub) HandleRequest(ctx context.Context, s rpc.Stream, req *rpc.Request) {
	m.pubsub.Publish(
		Request{
			Ctx:    ctx,
			Stream: s,
			Req:    req,
		},
	)
}

func (m *RequestPubSub) SubscribeToRequests(ctx context.Context) <-chan Request {
	return m.pubsub.Subscribe(ctx)
}
