package rooms_test

import (
	"context"
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/rooms/features"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/stretchr/testify/require"
)

func TestScanner_RunSimplyExitsIfTunnelingIsNotSupported(t *testing.T) {
	ts := newTestScanner(t)

	ts.MetadataGetter.GetMetadataValue = messages.NewRoomMetadataResponse(
		fixtures.SomeBool(),
		features.Features{},
	)

	ctx := fixtures.TestContext(t)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), mocks.NewConnectionMock(ctx))

	err := ts.Scanner.Run(ctx, peer)
	require.NoError(t, err)

	require.Equal(t, []transport.Peer{peer}, ts.MetadataGetter.GetMetadataCalls)
	require.Empty(t, ts.AttendantsGetter.GetAttendantsCalls)
}

func TestScanner_RunSimplyGetsAttendantsAndPublishesEventsIfTunnelingIsSupported(t *testing.T) {
	ts := newTestScanner(t)

	ts.MetadataGetter.GetMetadataValue = messages.NewRoomMetadataResponse(
		fixtures.SomeBool(),
		features.MustNewFeatures([]features.Feature{features.FeatureTunnel}),
	)

	event1 := rooms.MustNewRoomAttendantsEvent(rooms.RoomAttendantsEventTypeJoined, fixtures.SomeRefIdentity())
	event2 := rooms.MustNewRoomAttendantsEvent(rooms.RoomAttendantsEventTypeJoined, fixtures.SomeRefIdentity())

	ts.AttendantsGetter.GetAttendantsValues = []rooms.RoomAttendantsEvent{
		event1,
		event2,
	}

	ctx := fixtures.TestContext(t)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), mocks.NewConnectionMock(ctx))

	err := ts.Scanner.Run(ctx, peer)
	require.NoError(t, err)

	require.Equal(t, []transport.Peer{peer}, ts.MetadataGetter.GetMetadataCalls)
	require.Equal(t, []transport.Peer{peer}, ts.AttendantsGetter.GetAttendantsCalls)
	require.Equal(t,
		[]publishAttendantEventCall{
			{
				Portal: peer,
				Event:  event1,
			},
			{
				Portal: peer,
				Event:  event2,
			},
		},
		ts.Publisher.PublishAttendantEventCalls,
	)
}

type testScanner struct {
	Scanner          *rooms.Scanner
	MetadataGetter   *metadataGetterMock
	AttendantsGetter *attendantsGetterMock
	Publisher        *attendantEventPublisherMock
}

func newTestScanner(t *testing.T) testScanner {
	metadataGetter := newMetadataGetterMock()
	attendantsGetter := newAttendantsGetterMock()
	publisher := newAttendantEventPublisherMock()
	logger := fixtures.TestLogger(t)
	scanner := rooms.NewScanner(metadataGetter, attendantsGetter, publisher, logger)

	return testScanner{
		Scanner:          scanner,
		MetadataGetter:   metadataGetter,
		AttendantsGetter: attendantsGetter,
		Publisher:        publisher,
	}
}

type metadataGetterMock struct {
	GetMetadataCalls []transport.Peer
	GetMetadataValue messages.RoomMetadataResponse
}

func newMetadataGetterMock() *metadataGetterMock {
	return &metadataGetterMock{}
}

func (m *metadataGetterMock) GetMetadata(ctx context.Context, peer transport.Peer) (messages.RoomMetadataResponse, error) {
	m.GetMetadataCalls = append(m.GetMetadataCalls, peer)
	return m.GetMetadataValue, nil
}

type attendantsGetterMock struct {
	GetAttendantsCalls  []transport.Peer
	GetAttendantsValues []rooms.RoomAttendantsEvent
}

func newAttendantsGetterMock() *attendantsGetterMock {
	return &attendantsGetterMock{}
}

func (a *attendantsGetterMock) GetAttendants(ctx context.Context, peer transport.Peer) (<-chan rooms.RoomAttendantsEvent, error) {
	a.GetAttendantsCalls = append(a.GetAttendantsCalls, peer)
	ch := make(chan rooms.RoomAttendantsEvent)
	go func() {
		defer close(ch)
		for _, value := range a.GetAttendantsValues {
			select {
			case <-ctx.Done():
				return
			case ch <- value:
				continue
			}
		}
	}()
	return ch, nil
}

type attendantEventPublisherMock struct {
	PublishAttendantEventCalls []publishAttendantEventCall
}

func newAttendantEventPublisherMock() *attendantEventPublisherMock {
	return &attendantEventPublisherMock{}
}

func (a *attendantEventPublisherMock) PublishAttendantEvent(ctx context.Context, portal transport.Peer, event rooms.RoomAttendantsEvent) error {
	a.PublishAttendantEventCalls = append(a.PublishAttendantEventCalls, publishAttendantEventCall{
		Portal: portal,
		Event:  event,
	})
	return nil
}

type publishAttendantEventCall struct {
	Portal transport.Peer
	Event  rooms.RoomAttendantsEvent
}
