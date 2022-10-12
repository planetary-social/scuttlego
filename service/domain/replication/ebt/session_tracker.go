package ebt

import (
	"context"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type SessionTracker struct {
	lock     sync.Mutex // secures sessions and waiting
	sessions internal.Set[rpc.ConnectionId]
	waiting  map[rpc.ConnectionId][]chan<- bool
}

func NewSessionTracker() *SessionTracker {
	return &SessionTracker{
		sessions: internal.NewSet[rpc.ConnectionId](),
		waiting:  make(map[rpc.ConnectionId][]chan<- bool),
	}
}

type SessionEndedFn func()

func (t *SessionTracker) OpenSession(id rpc.ConnectionId) (SessionEndedFn, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.sessions.Contains(id) {
		return nil, errors.New("session already started")
	}

	t.sessions.Put(id)

	return t.sessionEndedFnForConn(id), nil
}

func (t *SessionTracker) WaitForSession(ctx context.Context, id rpc.ConnectionId, waitTime time.Duration) bool {
	ch := make(chan bool)
	t.registerWaitChannel(id, ch)
	go t.closeChannelAfterWaitTimeIfSessionDoesNotExist(id, ch, waitTime)

	select {
	case v := <-ch:
		return v
	case <-ctx.Done():
		return false
	}
}

func (t *SessionTracker) SomeoneIsWaiting(id rpc.ConnectionId) bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	return len(t.waiting[id]) > 0
}

func (t *SessionTracker) registerWaitChannel(id rpc.ConnectionId, ch chan<- bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.waiting[id] = append(t.waiting[id], ch)
}

func (t *SessionTracker) sessionEndedFnForConn(id rpc.ConnectionId) SessionEndedFn {
	return func() {
		t.lock.Lock()
		defer t.lock.Unlock()

		t.sessions.Delete(id)
		for _, channel := range t.waiting[id] {
			channel <- true
			close(channel)
		}
		delete(t.waiting, id)
	}
}

func (t *SessionTracker) closeChannelAfterWaitTimeIfSessionDoesNotExist(id rpc.ConnectionId, ch chan bool, waitTime time.Duration) {
	<-time.After(waitTime)

	t.lock.Lock()
	defer t.lock.Unlock()

	if !t.sessions.Contains(id) {
		for i, currentChannel := range t.waiting[id] {
			if currentChannel == ch {
				currentChannel <- false
				close(currentChannel)
				t.waiting[id] = append(t.waiting[id][:i], t.waiting[id][i+1:]...)
				break
			}
		}
		if len(t.waiting[id]) == 0 {
			delete(t.waiting, id)
		}
	}
}
