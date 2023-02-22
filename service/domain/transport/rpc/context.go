package rpc

import (
	"context"

	"github.com/planetary-social/scuttlego/service/domain/identity"
)

type connectionIdKeyType string

const connectionIdKey connectionIdKeyType = "connection_id"

func PutConnectionIdInContext(ctx context.Context, id ConnectionId) context.Context {
	return context.WithValue(ctx, connectionIdKey, id)
}

func GetConnectionIdFromContext(ctx context.Context) (ConnectionId, bool) {
	v := ctx.Value(connectionIdKey)
	if v == nil {
		return ConnectionId{}, false
	}
	return v.(ConnectionId), true
}

type remoteIdentityKeyType string

const remoteIdentityKey remoteIdentityKeyType = "connection_id"

func PutRemoteIdentityInContext(ctx context.Context, remoteIdentity identity.Public) context.Context {
	return context.WithValue(ctx, remoteIdentityKey, remoteIdentity)
}

func GetRemoteIdentityFromContext(ctx context.Context) (identity.Public, bool) {
	v := ctx.Value(remoteIdentityKey)
	if v == nil {
		return identity.Public{}, false
	}
	return v.(identity.Public), true
}
