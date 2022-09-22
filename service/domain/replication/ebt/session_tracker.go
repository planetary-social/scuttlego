package ebt

import (
	"context"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

const (
	waitForRemoteToStartEbtSessionFor = 5 * time.Second
)

type SessionTracker struct {
	lock     sync.Mutex // secures sessions and waiting
	sessions internal.Set[rpc.ConnectionId]
	waiting  map[rpc.ConnectionId][]chan<- struct{}
}

func NewSessionTracker() *SessionTracker {
	return &SessionTracker{
		sessions: internal.NewSet[rpc.ConnectionId](),
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

// WaitForSession waits for a session for the given connection to start and then
// blocks as long as this session is active. If a session doesn't start for a
// certain amount of time an error is returned. If the session terminates an
// error is returned.
func (t *SessionTracker) WaitForSession(ctx context.Context, id rpc.ConnectionId) error {
	select {
	case <-time.After(waitForRemoteToStartEbtSessionFor):
	case <-ctx.Done():
		return ctx.Err()
	}

	ch := make(chan struct{})

	if err := t.registerWaitChannelIfSessionExists(id, ch); err != nil {
		return errors.Wrap(err, "error registering the wait channel")
	}

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *SessionTracker) registerWaitChannelIfSessionExists(id rpc.ConnectionId, ch chan<- struct{}) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if !t.sessions.Contains(id) {
		return errors.New("session wasn't started")
	}

	t.waiting[id] = append(t.waiting[id], ch)
	return nil
}

func (t *SessionTracker) sessionEndedFnForConn(id rpc.ConnectionId) SessionEndedFn {
	return func() {
		t.lock.Lock()
		defer t.lock.Unlock()

		t.sessions.Delete(id)
		for _, channel := range t.waiting[id] {
			close(channel)
		}
		delete(t.waiting, id)
	}
}
