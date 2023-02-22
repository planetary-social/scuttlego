// Package transport implements the protocol stack responsible for exchanging
// data between Secure Scuttlebutt peers.
package transport

import (
	"context"
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

// Connection represents an RPC connection to a peer.
type Connection interface {
	PerformRequest(ctx context.Context, req *rpc.Request) (rpc.ResponseStream, error)

	// WasInitiatedByRemote returns true if this is a connection that was
	// initiated by the remote peer.
	WasInitiatedByRemote() bool

	Close() error
}

// Peer exists just for the purpose of keeping track of a connection together
// with the identity of the remote node.
//
// In theory that identity could be kept at the connection level and returned by
// the Connection interface however at the protocol level the concept of a
// remote identity exists only during the handshake. According to the protocol
// the RPC connection itself doesn't really know about the handshake or the
// identity. Those are properties of the underlying boxstream transport layer.
type Peer struct {
	remote identity.Public
	conn   Connection
}

func NewPeer(remote identity.Public, conn Connection) (Peer, error) {
	if remote.IsZero() {
		return Peer{}, errors.New("zero value of remote identity")
	}

	if conn == nil {
		return Peer{}, errors.New("conn is nil")
	}

	return Peer{
		remote: remote,
		conn:   conn,
	}, nil
}

func MustNewPeer(remote identity.Public, conn Connection) Peer {
	v, err := NewPeer(remote, conn)
	if err != nil {
		panic(err)
	}
	return v
}

func (p Peer) Identity() identity.Public {
	return p.remote
}

func (p Peer) Conn() Connection {
	return p.conn
}

func (p Peer) IsZero() bool {
	return p.conn == nil
}

func (p Peer) String() string {
	public, _ := refs.NewIdentityFromPublic(p.remote)
	return fmt.Sprintf("<peer identity=%s conn=%v>", public.String(), p.conn)
}
