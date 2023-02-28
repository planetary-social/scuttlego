package ebt

import (
	"context"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/internal/fixtures"
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
	require.NotPanics(t, func() {
		done()
	})
}

func TestSessionTracker_WaitForSessionExitsAndReturnsFalseIfSessionDoesNotExist(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	ctx := fixtures.TestContext(t)

	ok, err := tracker.WaitForSession(ctx, fixtures.SomeConnectionId(), testWaitDuration)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestSessionTracker_WaitForSessionExitsAndReturnsTrueIfSessionExistsButTerminatesAfterGracePeriod_OpenFirst(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	ctx := fixtures.TestContext(t)
	connectionId := fixtures.SomeConnectionId()

	doneCh := make(chan bool)
	var doneDuration time.Duration

	done, err := tracker.OpenSession(connectionId)
	require.NoError(t, err)

	go func() {
		defer close(doneCh)

		start := time.Now()
		ok, err := tracker.WaitForSession(ctx, connectionId, testWaitDuration)
		require.NoError(t, err)
		doneDuration = time.Since(start)
		select {
		case <-ctx.Done():
		case doneCh <- ok:
		}
	}()

	<-time.After(2 * testWaitDuration)
	done()

	select {
	case result, ok := <-doneCh:
		require.True(t, ok)
		require.True(t, result)
		require.NotZero(t, doneDuration)
		require.Greater(t, doneDuration, 2*testWaitDuration, "wait for session exited before the session was marked as done (probably due to the timer)")
		require.False(t, tracker.SomeoneIsWaiting(connectionId))
	case <-time.After(4 * testWaitDuration):
		t.Fatal("timeout")
	}
}

func TestSessionTracker_WaitForSessionExitsAndReturnsTrueIfSessionExistsButTerminatesBeforeGracePeriod_OpenFirst(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	ctx := fixtures.TestContext(t)
	connectionId := fixtures.SomeConnectionId()

	doneCh := make(chan bool)
	var doneDuration time.Duration

	done, err := tracker.OpenSession(connectionId)
	require.NoError(t, err)

	go func() {
		defer close(doneCh)

		start := time.Now()
		ok, err := tracker.WaitForSession(ctx, connectionId, testWaitDuration)
		require.NoError(t, err)
		doneDuration = time.Since(start)
		select {
		case <-ctx.Done():
		case doneCh <- ok:
		}
	}()

	<-time.After(testWaitDuration / 2)
	done()

	select {
	case result, ok := <-doneCh:
		require.True(t, ok)
		require.True(t, result)
		require.NotZero(t, doneDuration)
		require.Less(t, doneDuration, testWaitDuration, "wait for session exited way after the session was marked as done (probably due to the timer)")
		require.False(t, tracker.SomeoneIsWaiting(connectionId))
	case <-time.After(4 * testWaitDuration):
		t.Fatal("timeout")
	}
}

func TestSessionTracker_WaitForSessionExitsAndReturnsTrueIfSessionExistsButTerminatesAfterGracePeriod_WaitFirst(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	ctx := fixtures.TestContext(t)
	connectionId := fixtures.SomeConnectionId()

	doneCh := make(chan bool)
	var doneDuration time.Duration

	go func() {
		defer close(doneCh)

		start := time.Now()
		ok, err := tracker.WaitForSession(ctx, connectionId, testWaitDuration)
		require.NoError(t, err)
		doneDuration = time.Since(start)
		select {
		case <-ctx.Done():
		case doneCh <- ok:
		}
	}()

	go func() {
		for {
			if tracker.SomeoneIsWaiting(connectionId) {
				break
			}
			<-time.After(testWaitDuration / 100)
		}

		done, err := tracker.OpenSession(connectionId)
		require.NoError(t, err)

		<-time.After(2 * testWaitDuration)
		done()
	}()

	select {
	case result, ok := <-doneCh:
		require.True(t, ok)
		require.True(t, result)
		require.NotZero(t, doneDuration)
		require.Greater(t, doneDuration, 2*testWaitDuration, "wait for session exited before the session was marked as done (probably due to the timer)")
	case <-time.After(4 * testWaitDuration):
		t.Fatal("timeout")
	}
}

func TestSessionTracker_WaitForSessionExitsAndReturnsTrueIfSessionExistsButTerminatesBeforeGracePeriod_WaitFirst(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	ctx := fixtures.TestContext(t)
	connectionId := fixtures.SomeConnectionId()

	doneCh := make(chan bool)
	var doneDuration time.Duration

	go func() {
		defer close(doneCh)

		start := time.Now()
		ok, err := tracker.WaitForSession(ctx, connectionId, testWaitDuration)
		require.NoError(t, err)
		doneDuration = time.Since(start)
		select {
		case <-ctx.Done():
		case doneCh <- ok:
		}
	}()

	go func() {
		for {
			if tracker.SomeoneIsWaiting(connectionId) {
				break
			}
			<-time.After(testWaitDuration / 100)
		}

		done, err := tracker.OpenSession(connectionId)
		require.NoError(t, err)

		<-time.After(testWaitDuration / 2)
		done()
	}()

	select {
	case result, ok := <-doneCh:
		require.True(t, ok)
		require.True(t, result)
		require.NotZero(t, doneDuration)
		require.Less(t, doneDuration, testWaitDuration, "wait for session exited way after the session was marked as done (probably due to the timer)")
	case <-time.After(4 * testWaitDuration):
		t.Fatal("timeout")
	}
}

func TestSessionTracker_WaitForSessionBlocksIfSessionExists(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	ctx := fixtures.TestContext(t)
	connectionId := fixtures.SomeConnectionId()
	doneCh := make(chan bool)

	go func() {
		ok, err := tracker.WaitForSession(ctx, connectionId, testWaitDuration)
		require.NoError(t, err)
		doneCh <- ok
	}()

	done, err := tracker.OpenSession(connectionId)
	require.NoError(t, err)

	select {
	case <-doneCh:
		t.Fatal("stopped blocking prematurely")
	case <-time.After(2 * testWaitDuration):
		t.Log("wait for session blocked correctly")
	}

	done()
	require.True(t, <-doneCh)
}

func TestSessionTracker_WaitForSessionExitsAndReturnsFalseIfSessionIsNeverCreated(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	ctx := fixtures.TestContext(t)
	connectionId := fixtures.SomeConnectionId()

	start := time.Now()
	ok, err := tracker.WaitForSession(ctx, connectionId, testWaitDuration)
	require.Greater(t, time.Since(start), testWaitDuration)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestSessionTracker_ErrIsReturnedIfContextTerminates(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	ctx, cancel := context.WithCancel(fixtures.TestContext(t))
	connectionId := fixtures.SomeConnectionId()

	go func() {
		<-time.After(testWaitDuration / 2)
		cancel()
	}()

	_, err := tracker.WaitForSession(ctx, connectionId, testWaitDuration)
	require.ErrorIs(t, err, context.Canceled)
}
