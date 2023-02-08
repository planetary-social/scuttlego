package rpc

import (
	"sync/atomic"
)

type ConnectionIdGenerator struct {
	lastId uint64
}

func NewConnectionIdGenerator() *ConnectionIdGenerator {
	return &ConnectionIdGenerator{}
}

func (v *ConnectionIdGenerator) Generate() ConnectionId {
	newId := atomic.AddUint64(&v.lastId, 1)
	return NewConnectionId(int(newId))
}
