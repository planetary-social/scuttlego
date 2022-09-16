package ebt

import (
	"context"
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"sync"
	"time"
)

const (
	waitForRemoteToStartEbtSessionFor = 5 * time.Second
)

type SessionTracker struct {
	lock     sync.Mutex // secures sessions and waiting
	sessions internal.Set[rpc.ConnectionId]
	waiting  map[rpc.ConnectionId][]chan error
}

func NewSessionTracker() *SessionTracker {
	return &SessionTracker{
		sessions: internal.NewSet[rpc.ConnectionId](),
	}
}

type SessionEndedFn func()

func (t *SessionTracker) OpenLocalSession(id rpc.ConnectionId) (SessionEndedFn, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.sessions.Contains(id) {
		return nil, errors.New("session already started")
	}

	return nil
}

func (t *SessionTracker) RemoteSessionOpened(id rpc.ConnectionId) (SessionEndedFn, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	return errors.New("not implemented")
}

func (t *SessionTracker) sessionEndedFnForConn(id rpc.ConnectionId) (SessionEndedFn) {
	return func() {
		t.lock.Lock()
		defer t.lock.Unlock()

		t.sessions.Delete(id)
	}
}

// WaitForSession waits for a session for the given connection to start and then
// blocks as long as this session is active. If a session doesn't start for a
// certain amount of time an error is returned. If the session terminates an
// error is returned.
func (t *SessionTracker) WaitForSession(ctx context.Context, id rpc.ConnectionId) error {
	ch := make(chan error)

	t.lock.Lock()
	defer t.lock.Unlock()

	waiting[id]

	select {
	case <-time.After(waitForRemoteToStartEbtSessionFor):
	case <-ctx.Done():
		return ctx.Err()
	}

	return ch
}
