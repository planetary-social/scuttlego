package commands_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/stretchr/testify/require"
)

func TestNewProcessRoomAttendantEvent_ValuesAreCorrectlySetAndReturnedByGetters(t *testing.T) {
	ctx := fixtures.TestContext(t)
	conn := mocks.NewConnectionMock(ctx)

	portal, err := transport.NewPeer(fixtures.SomePublicIdentity(), conn)
	require.NoError(t, err)

	event, err := rooms.NewRoomAttendantsEvent(rooms.RoomAttendantsEventTypeJoined, fixtures.SomeRefIdentity())
	require.NoError(t, err)

	cmd, err := commands.NewProcessRoomAttendantEvent(portal, event)
	require.NoError(t, err)

	require.Equal(t, portal, cmd.Portal())
	require.Equal(t, event, cmd.Event())
	require.False(t, cmd.IsZero())
}

func TestProcessRoomAttendantEventHandler_ReturnsErrorOnZeroValueOfCommand(t *testing.T) {
	tc, err := di.BuildTestCommands(t)
	require.NoError(t, err)

	ctx := fixtures.TestContext(t)

	err = tc.ProcessRoomAttendantEvent.Handle(ctx, commands.ProcessRoomAttendantEvent{})
	require.EqualError(t, err, "zero value of command")
}

func TestProcessRoomAttendantEventHandler_CallsPeerManagerConnectViaRoomForSpecificEventTypes(t *testing.T) {
	testCases := []struct {
		Name                     string
		EventType                rooms.RoomAttendantsEventType
		ShouldCallConnectViaRoom bool
	}{
		{
			Name:                     "joined",
			EventType:                rooms.RoomAttendantsEventTypeJoined,
			ShouldCallConnectViaRoom: true,
		},
		{
			Name:                     "left",
			EventType:                rooms.RoomAttendantsEventTypeLeft,
			ShouldCallConnectViaRoom: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tc, err := di.BuildTestCommands(t)
			require.NoError(t, err)

			ctx := fixtures.TestContext(t)
			conn := mocks.NewConnectionMock(ctx)

			portal, err := transport.NewPeer(fixtures.SomePublicIdentity(), conn)
			require.NoError(t, err)

			target := fixtures.SomeRefIdentity()

			event, err := rooms.NewRoomAttendantsEvent(testCase.EventType, target)
			require.NoError(t, err)

			cmd, err := commands.NewProcessRoomAttendantEvent(portal, event)
			require.NoError(t, err)

			err = tc.ProcessRoomAttendantEvent.Handle(ctx, cmd)
			require.NoError(t, err)

			if testCase.ShouldCallConnectViaRoom {
				require.Equal(t,
					[]mocks.PeerManagerConnectViaRoomCall{
						{
							Portal: portal,
							Target: target.Identity(),
						},
					},
					tc.PeerManager.ConnectViaRoomCalls(),
				)
			} else {
				require.Empty(t, tc.PeerManager.ConnectViaRoomCalls())
			}
		})
	}
}
