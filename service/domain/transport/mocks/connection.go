package mocks

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type ConnectionMock struct {
	ctx    context.Context
	cancel context.CancelFunc

	msgFn func(req *rpc.Request) []rpc.ResponseWithError
}

func NewConnectionMock(ctx context.Context) *ConnectionMock {
	ctx, cancel := context.WithCancel(ctx)
	return &ConnectionMock{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *ConnectionMock) WasInitiatedByRemote() bool {
	return fixtures.SomeBool()
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
	ch chan rpc.ResponseWithError
}

func newResponseStreamMock(ch chan rpc.ResponseWithError) *responseStreamMock {
	return &responseStreamMock{
		ch: ch,
	}
}

func (r responseStreamMock) WriteMessage(body []byte) error {
	return errors.New("not implemented")
}

func (r responseStreamMock) Channel() <-chan rpc.ResponseWithError {
	return r.ch
}
