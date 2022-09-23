package ebt

import (
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/stretchr/testify/require"
)

const (
	testWaitDuration = 100 * time.Millisecond
)

func TestSessionTracker_OpenSessionReturnsAnInitializedDoneFunction(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()

	done, err := tracker.OpenSession(fixtures.SomeConnectionId())
	require.NoError(t, err)
	require.NotNil(t, done)
}

func TestSessionTracker_WaitForSessionExitsAndReturnsAnErrorIfSessionDoesNotExist(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	ctx := fixtures.TestContext(t)

	err := tracker.WaitForSession(ctx, fixtures.SomeConnectionId(), testWaitDuration)
	require.Error(t, err)
}

func TestSessionTracker_WaitForSessionExitsAndDoesNotReturnAnErrorIfSessionExists(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	ctx := fixtures.TestContext(t)
	connectionId := fixtures.SomeConnectionId()
	doneCh := make(chan error)

	done, err := tracker.OpenSession(connectionId)
	require.NoError(t, err)

	go func() {
		defer close(doneCh)

		err := tracker.WaitForSession(ctx, connectionId, testWaitDuration)
		select {
		case <-ctx.Done():
		case doneCh <- err:
		}
	}()

	<-time.After(2 * testWaitDuration)
	done()

	select {
	case err := <-doneCh:
		require.NoError(t, err)
	case <-time.After(4 * testWaitDuration):
		t.Fatal("timeout")
	}
}

func TestSessionTracker_WaitForSessionBlocksIfSessionExists(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	ctx := fixtures.TestContext(t)
	connectionId := fixtures.SomeConnectionId()
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)

		tracker.WaitForSession(ctx, connectionId, testWaitDuration)
	}()

	done, err := tracker.OpenSession(connectionId)
	require.NoError(t, err)
	defer done()

	select {
	case <-doneCh:
		t.Fatal("stopped blocking prematurely")
	case <-time.After(2 * testWaitDuration):
		t.Log("wait for session blocked correctly")
	}
}
