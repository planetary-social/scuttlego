package ebt_test

import (
	"context"
	"testing"

	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/stretchr/testify/require"
)

func TestRequestedFeeds_RequestingCallsMessageStreamerAndRequestingAgainCancelsThePreviousCall(t *testing.T) {
	stream := newMockStream()
	messageStreamer := newMessageStreamerMock()

	ctx := fixtures.TestContext(t)
	ref := fixtures.SomeRefFeed()
	seq := internal.Ptr(fixtures.SomeSequence())

	rf := ebt.NewRequestedFeeds(messageStreamer, stream)
	rf.Request(ctx, ref, seq)

	require.Len(t, messageStreamer.Calls, 1)
	require.Equal(t, ref, messageStreamer.Calls[0].Id)
	require.Equal(t, seq, messageStreamer.Calls[0].Seq)

	rf.Request(ctx, ref, seq)

	require.Len(t, messageStreamer.Calls, 2)
	require.Equal(t, ref, messageStreamer.Calls[1].Id)
	require.Equal(t, seq, messageStreamer.Calls[1].Seq)

	select {
	case <-messageStreamer.Calls[0].Ctx.Done():
		t.Log("context is cancelled as expected")
	default:
		t.Fatal("timeout, context is not cancelled")
	}

	select {
	case <-messageStreamer.Calls[1].Ctx.Done():
		t.Fatal("context was not expected to be cancelled")
	default:
		t.Log("context isn't cancelled as expected")
	}
}

func TestRequestedFeeds_CancellingIfNotRequestedFirstDoesNothing(t *testing.T) {
	stream := newMockStream()
	messageStreamer := newMessageStreamerMock()

	ref := fixtures.SomeRefFeed()

	rf := ebt.NewRequestedFeeds(messageStreamer, stream)
	rf.Cancel(ref)
}

func TestRequestedFeeds_CancellingCancellsPreviousRequest(t *testing.T) {
	stream := newMockStream()
	messageStreamer := newMessageStreamerMock()

	ctx := fixtures.TestContext(t)
	ref := fixtures.SomeRefFeed()
	seq := internal.Ptr(fixtures.SomeSequence())

	rf := ebt.NewRequestedFeeds(messageStreamer, stream)
	rf.Request(ctx, ref, seq)

	require.Len(t, messageStreamer.Calls, 1)

	select {
	case <-messageStreamer.Calls[0].Ctx.Done():
		t.Fatal("context was cancelled for some reason")
	default:
		t.Log("ok, context isn't cancelled")
	}

	rf.Cancel(ref)

	select {
	case <-messageStreamer.Calls[0].Ctx.Done():
		t.Log("ok, context is cancelled as expected")
	default:
		t.Fatal("timeout, context is not cancelled")
	}
}

type messageStreamerMock struct {
	Calls []messageStreamerCall
}

func newMessageStreamerMock() *messageStreamerMock {
	return &messageStreamerMock{}
}

func (m *messageStreamerMock) Handle(ctx context.Context, id refs.Feed, seq *message.Sequence, messageWriter ebt.MessageWriter) {
	m.Calls = append(m.Calls, messageStreamerCall{
		Ctx: ctx,
		Id:  id,
		Seq: seq,
	})
}

type messageStreamerCall struct {
	Ctx context.Context
	Id  refs.Feed
	Seq *message.Sequence
}
