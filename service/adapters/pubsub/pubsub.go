package pubsub

import (
	"context"
	"sync"
)

type channelWithContext[T any] struct {
	Ch  chan T
	Ctx context.Context
}

type GoChannelPubSub[T any] struct {
	subscriptions []channelWithContext[T]
	lock          sync.Mutex
}

func NewGoChannelPubSub[T any]() *GoChannelPubSub[T] {
	return &GoChannelPubSub[T]{}
}

func (g *GoChannelPubSub[T]) Subscribe(ctx context.Context) <-chan T {
	ch := make(chan T)

	g.addSubscription(ctx, ch)

	go func() {
		<-ctx.Done()
		g.removeSubscription(ch)
		close(ch)
	}()

	return ch
}

func (g *GoChannelPubSub[T]) Publish(value T) {
	g.lock.Lock()
	defer g.lock.Unlock()

	for _, sub := range g.subscriptions {
		select {
		case sub.Ch <- value:
		case <-sub.Ctx.Done():
		}
	}
}

func (g *GoChannelPubSub[T]) addSubscription(ctx context.Context, ch chan T) {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.subscriptions = append(g.subscriptions, channelWithContext[T]{Ch: ch, Ctx: ctx})
}

func (g *GoChannelPubSub[T]) removeSubscription(ch chan T) {
	g.lock.Lock()
	defer g.lock.Unlock()

	for i := range g.subscriptions {
		if g.subscriptions[i].Ch == ch {
			g.subscriptions = append(g.subscriptions[:i], g.subscriptions[i+1:]...)
			return
		}
	}

	panic("somehow the subscription was already removed, this must be a bug")
}
