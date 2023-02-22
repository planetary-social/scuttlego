package domain

import (
	"context"
	"encoding/base64"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type Dialer interface {
	Dial(ctx context.Context, remote identity.Public, address network.Address) (transport.Peer, error)
}

type RoomDialer interface {
	DialViaRoom(ctx context.Context, portal transport.Peer, target identity.Public) (transport.Peer, error)
}

type PeerManagerConfig struct {
	// Peer manager will attempt to remain connected to the preferred pubs.
	PreferredPubs []Pub
}

type Pub struct {
	Identity identity.Public
	Address  network.Address
}

type PeerManager struct {
	peers     peersMap
	peersLock *sync.Mutex

	config PeerManagerConfig

	dialer     Dialer
	roomDialer RoomDialer
	logger     logging.Logger
}

// NewPeerManager creates a new peer manager. The provided context is used as the base context for RPC connections
// established by the manager.
func NewPeerManager(
	config PeerManagerConfig,
	dialer Dialer,
	roomDialer RoomDialer,
	logger logging.Logger,
) *PeerManager {
	return &PeerManager{
		peers:      make(peersMap),
		peersLock:  &sync.Mutex{},
		config:     config,
		dialer:     dialer,
		roomDialer: roomDialer,
		logger:     logger.New("peer_manager"),
	}
}

// EstablishNewConnections tries to establish new connections to the preferred pubs.
func (p PeerManager) EstablishNewConnections(ctx context.Context) error {
	var resultErr error

	for _, pub := range p.config.PreferredPubs {
		if err := p.Connect(ctx, pub.Identity, pub.Address); err != nil {
			resultErr = multierror.Append(
				resultErr,
				errors.Wrapf(err, "failed to connect to to a pub '%s' on '%s'", pub.Identity, pub.Address),
			)
		}
	}

	return resultErr
}

// Peers returns a list of peers currently tracked by the manager.
func (p PeerManager) Peers() []transport.Peer {
	p.peersLock.Lock()
	defer p.peersLock.Unlock()

	var result []transport.Peer
	for _, connectedPeer := range p.peers {
		result = append(result, connectedPeer.peer)
	}
	return result
}

// DisconnectAll disconnects all peers.
func (p PeerManager) DisconnectAll() error {
	p.peersLock.Lock()
	defer p.peersLock.Unlock()

	var resultErr error
	for _, connectedPeer := range p.peers {
		if err := connectedPeer.peer.Conn().Close(); err != nil {
			resultErr = multierror.Append(resultErr, err)
		}
	}
	return resultErr
}

// Connect attempts to establish communications with the specified peer. If a
// connection to the specified peer already exists then a new connection will
// not be initiated. If connecting to the peer succeeds but in the meantime a
// connection to the same node was created manually or automatically by the
// manager then the old connection will be replaced by the new connection and
// terminated.
func (p PeerManager) Connect(ctx context.Context, remote identity.Public, address network.Address) error {
	if p.alreadyConnected(remote) { // early check
		return nil
	}

	p.logger.WithField("remote", remote).WithField("address", address).Debug("dialing")

	_, err := p.dialer.Dial(ctx, remote, address)
	if err != nil {
		return errors.Wrap(err, "dial failed")
	}

	return nil
}

// ConnectViaRoom attempts to establish communications with the specified peer
// using a room as a relay. Behaves like Connect.
func (p PeerManager) ConnectViaRoom(ctx context.Context, portal transport.Peer, target identity.Public) error {
	if p.alreadyConnected(target) { // early check
		return nil
	}

	p.logger.WithField("target", target).WithField("portal", portal).Debug("dialing via room")

	_, err := p.roomDialer.DialViaRoom(ctx, portal, target)
	if err != nil {
		return errors.Wrap(err, "dial via room failed")
	}

	return nil
}

// ProcessNewLocalDiscovery handles incoming local peer announcements.
func (p PeerManager) ProcessNewLocalDiscovery(ctx context.Context, remote identity.Public, address network.Address) error {
	return p.Connect(ctx, remote, address)
}

func (p PeerManager) TrackPeer(ctx context.Context, peer transport.Peer) {
	p.peersLock.Lock()
	defer p.peersLock.Unlock()

	key := p.peerKey(peer.Identity())

	existingPeer, ok := p.peers[key]
	if ok {
		go func() {
			existingPeer.peer.Conn().Close()
		}()
	}

	p.peers[key] = newConnectedPeer(peer)
	go p.removePeerIfConnectionCloses(ctx, peer)
}

func (p PeerManager) removePeerIfConnectionCloses(ctx context.Context, peer transport.Peer) {
	<-ctx.Done()

	p.peersLock.Lock()
	defer p.peersLock.Unlock()

	key := p.peerKey(peer.Identity())

	trackedPeer, ok := p.peers[key]
	if ok && trackedPeer.peer.Conn() == peer.Conn() {
		delete(p.peers, key)
	}
}

func (p PeerManager) alreadyConnected(remote identity.Public) bool {
	p.peersLock.Lock()
	defer p.peersLock.Unlock()

	_, ok := p.peers[p.peerKey(remote)]
	return ok
}

func (p PeerManager) peerKey(remote identity.Public) string {
	return base64.StdEncoding.EncodeToString(remote.PublicKey())
}

type connectedPeer struct {
	peer  transport.Peer
	added time.Time
}

func newConnectedPeer(peer transport.Peer) connectedPeer {
	return connectedPeer{
		peer:  peer,
		added: time.Now(),
	}
}

type peersMap map[string]connectedPeer
