package rpc

import (
	"context"
	"strconv"
)

type ConnectionId struct {
	id int
}

func NewConnectionId(id int) ConnectionId {
	return ConnectionId{id: id}
}

func (c ConnectionId) String() string {
	return strconv.Itoa(c.id)
}

const connectionIdKey = "connection_id"

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
