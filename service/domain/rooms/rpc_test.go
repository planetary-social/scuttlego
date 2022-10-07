package rooms_test

import (
	"context"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/rooms/features"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestGetMetadata(t *testing.T) {
	ctx := fixtures.TestContext(t)

	conn := mocks.NewConnectionMock(ctx)
	peer := transport.NewPeer(fixtures.SomePublicIdentity(), conn)

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

	metadata, err := rooms.GetMetadata(ctx, peer)
	require.NoError(t, err)

	require.True(t, metadata.Membership())
	require.True(t, metadata.Features().Contains(features.FeatureTunnel))
}
