package rooms_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/rooms/features"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestGetMetadata(t *testing.T) {
	ctx := fixtures.TestContext(t)

	conn := mocks.NewConnectionMock(ctx)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), conn)

	conn.Mock(func(req *rpc.Request) []rpc.ResponseWithError {
		return []rpc.ResponseWithError{
			{
				Value: rpc.NewResponse([]byte(`{"membership": true, "features": ["tunnel", "some-other-feature"]}`)),
				Err:   nil,
			},
		}

	})

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	logger := fixtures.TestLogger(t)
	adapter := rooms.NewPeerRPCAdapter(logger)

	metadata, err := adapter.GetMetadata(ctx, peer)
	require.NoError(t, err)

	require.True(t, metadata.Membership())
	require.True(t, metadata.Features().Contains(features.FeatureTunnel))
}

func TestGetAttendants(t *testing.T) {
	ctx := fixtures.TestContext(t)

	conn := mocks.NewConnectionMock(ctx)
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), conn)

	ref1 := refs.MustNewIdentity("@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519")
	ref2 := refs.MustNewIdentity("@gYVa2GgdDYbR6R4AFnk5y2aU0sQirNIIoAcpOUh/aZk=.ed25519")
	ref3 := refs.MustNewIdentity("@650YpEeEBF2H88Z88idG6ZWvWiU2eVG6ov9s1HHEg/E=.ed25519")
	ref4 := refs.MustNewIdentity("@YyUlP+xzjdep4ov5IRGcFg8HAkSGFbvaCDE/ao62aNI=.ed25519")

	conn.Mock(func(req *rpc.Request) []rpc.ResponseWithError {
		return []rpc.ResponseWithError{
			{
				Value: rpc.NewResponse([]byte(fmt.Sprintf(`{"type": "state", "ids": ["%s", "%s"]}`, ref1, ref2))),
				Err:   nil,
			},
			{
				Value: rpc.NewResponse([]byte(fmt.Sprintf(`{"type": "joined", "id": "%s"}`, ref3))),
				Err:   nil,
			},
			{
				Value: rpc.NewResponse([]byte(fmt.Sprintf(`{"type": "left", "id": "%s"}`, ref4))),
				Err:   nil,
			},
		}

	})

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	logger := fixtures.TestLogger(t)
	adapter := rooms.NewPeerRPCAdapter(logger)

	ch, err := adapter.GetAttendants(ctx, peer)
	require.NoError(t, err)

	var result []rooms.RoomAttendantsEvent

	for v := range ch {
		result = append(result, v)
	}

	require.Equal(t,
		[]rooms.RoomAttendantsEvent{
			rooms.MustNewRoomAttendantsEvent(rooms.RoomAttendantsEventTypeJoined, ref1),
			rooms.MustNewRoomAttendantsEvent(rooms.RoomAttendantsEventTypeJoined, ref2),
			rooms.MustNewRoomAttendantsEvent(rooms.RoomAttendantsEventTypeJoined, ref3),
			rooms.MustNewRoomAttendantsEvent(rooms.RoomAttendantsEventTypeLeft, ref4),
		},
		result,
	)
}
