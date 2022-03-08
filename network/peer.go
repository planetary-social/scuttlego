package network

import (
	"github.com/planetary-social/go-ssb/identity"
	"github.com/planetary-social/go-ssb/network/rpc"
	"github.com/planetary-social/go-ssb/refs"
)

type Peer struct {
	remote identity.Public
	conn   *rpc.Connection
}

func NewPeer(remote identity.Public, conn *rpc.Connection) Peer {
	return Peer{
		remote: remote,
		conn:   conn,
	}
}

func (p Peer) Identity() identity.Public {
	return p.remote
}

func (p Peer) Conn() *rpc.Connection {
	return p.conn
}

func (p Peer) String() string {
	public, _ := refs.NewIdentityFromPublic(p.remote)
	return public.String()
}
