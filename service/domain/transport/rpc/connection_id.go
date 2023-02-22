package rpc

import (
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
