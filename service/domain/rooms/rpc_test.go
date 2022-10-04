package rooms_test

import (
	"context"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetMetadata(t *testing.T) {
	ctx := fixtures.TestContext(t)

	conn := mocks.NewConnectionMock(ctx)
	peer := transport.NewPeer(fixtures.SomeRefIdentity(), conn)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	metadata, err := rooms.GetMetadata(ctx, peer)
	require.NoError(t, err)
}
