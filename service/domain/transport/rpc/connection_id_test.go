package rpc_test

import (
	"context"
	"testing"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestConnectionIdContext(t *testing.T) {
	ctx := context.Background()
	connectionId := fixtures.SomeConnectionId()

	ctx = rpc.PutConnectionIdInContext(ctx, connectionId)

	retrievedId, ok := rpc.GetConnectionIdFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, connectionId, retrievedId)
}
