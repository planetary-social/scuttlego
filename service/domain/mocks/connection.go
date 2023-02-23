package mocks

import (
	"context"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
)

type ConnectionMock struct {
	ctx                  context.Context
	cancel               context.CancelFunc
	wasInitiatedByRemote bool

	msgFn func(req *rpc.Request) []rpc.ResponseWithError
}

func NewConnectionMock(ctx context.Context) *ConnectionMock {
	ctx, cancel := context.WithCancel(ctx)
	return &ConnectionMock{
		ctx:                  ctx,
		cancel:               cancel,
		wasInitiatedByRemote: fixtures.SomeBool(),
	}
}

func (c *ConnectionMock) WasInitiatedByRemote() bool {
	return c.wasInitiatedByRemote
}

func (c *ConnectionMock) SetWasInitiatedByRemote(wasInitiatedByRemote bool) {
	c.wasInitiatedByRemote = wasInitiatedByRemote
}

func (c *ConnectionMock) PerformRequest(ctx context.Context, req *rpc.Request) (rpc.ResponseStream, error) {
	ch := make(chan rpc.ResponseWithError)
	m := newResponseStreamMock(ch)

	go func() {
		defer close(ch)

		for _, msg := range c.msgFn(req) {
			select {
			case ch <- msg:
				continue
			case <-ctx.Done():
				return
			}
		}
	}()

	return m, nil
}

func (c *ConnectionMock) Context() context.Context {
	return c.ctx
}

func (c *ConnectionMock) Close() error {
	c.cancel()
	return nil
}

func (c *ConnectionMock) IsClosed() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}

func (c *ConnectionMock) Mock(fn func(req *rpc.Request) []rpc.ResponseWithError) {
	c.msgFn = fn
}

type responseStreamMock struct {
	chReceive chan rpc.ResponseWithError
}

func newResponseStreamMock(chReceive chan rpc.ResponseWithError) *responseStreamMock {
	return &responseStreamMock{
		chReceive: chReceive,
	}
}

func (r responseStreamMock) WriteMessage(body []byte, bodyType transport.MessageBodyType) error {
	return nil
}

func (r responseStreamMock) Channel() <-chan rpc.ResponseWithError {
	return r.chReceive
}

func (r responseStreamMock) Ctx() context.Context {
	return context.TODO()
}
