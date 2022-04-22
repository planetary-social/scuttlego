package transport

import (
	"fmt"

	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
)

// Peer is here just for the purpose of storing an RPC connection together with the identity of the remote node. In
// theory that identity could be placed inside the rpc.Connection struct however at the protocol level the concept
// of a remote identity exists only during the handshake, the RPC connection itself doesn't really know about the
// handshake or the identity. Those are properties of the underlying boxstream transport layer.
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
	return fmt.Sprintf("<peer %s>", public.String())
}
