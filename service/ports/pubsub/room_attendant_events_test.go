package pubsub_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	pubsub2 "github.com/planetary-social/scuttlego/service/adapters/pubsub"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/ports/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoomAttendantEventSubscriber_ReceivesEventsAndCallsTheCommandHandler(t *testing.T) {
	ctx := fixtures.TestContext(t)

	ps := pubsub2.NewRoomAttendantEventPubSub()
	logger := fixtures.TestLogger(t)
	handler := newProcessRoomAttendantEventHandlerMock()

	subscriber := pubsub.NewRoomAttendantEventSubscriber(
		ps,
		handler,
		logger,
	)
	go subscriber.Run(ctx) //nolint:errcheck

	portal := transport.MustNewPeer(fixtures.SomePublicIdentity(), mocks.NewConnectionMock(ctx))
	event, err := rooms.NewRoomAttendantsEvent(rooms.RoomAttendantsEventTypeJoined, fixtures.SomeRefIdentity())
	require.NoError(t, err)

	<-time.After(100 * time.Millisecond)

	err = ps.PublishAttendantEvent(ctx, portal, event)
	require.NoError(t, err)

	cmd, err := commands.NewProcessRoomAttendantEvent(portal, event)
	require.NoError(t, err)

	require.Eventually(t,
		func() bool {
			return assert.ObjectsAreEqual([]commands.ProcessRoomAttendantEvent{cmd}, handler.Calls())
		}, 1*time.Second, 100*time.Millisecond)
}

type processRoomAttendantEventHandlerMock struct {
	lock  sync.Mutex
	calls []commands.ProcessRoomAttendantEvent
}

func newProcessRoomAttendantEventHandlerMock() *processRoomAttendantEventHandlerMock {
	return &processRoomAttendantEventHandlerMock{}
}

func (p *processRoomAttendantEventHandlerMock) Handle(ctx context.Context, cmd commands.ProcessRoomAttendantEvent) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.calls = append(p.calls, cmd)
	return nil
}

func (p *processRoomAttendantEventHandlerMock) Calls() []commands.ProcessRoomAttendantEvent {
	p.lock.Lock()
	defer p.lock.Unlock()
	tmp := make([]commands.ProcessRoomAttendantEvent, len(p.calls))
	copy(tmp, p.calls)
	return tmp
}
