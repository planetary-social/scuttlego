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

type MessageReplicator interface {
	Replicate(ctx context.Context, peer transport.Peer) error
}

type BlobReplicator interface {
	Replicate(ctx context.Context, peer transport.Peer) error
}

type RoomScanner interface {
	Run(ctx context.Context, peer transport.Peer) error
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
	ctx context.Context

	peers     peersMap
	peersLock *sync.Mutex

	config PeerManagerConfig

	dialer            Dialer
	messageReplicator MessageReplicator
	blobReplicator    BlobReplicator
	roomScanner       RoomScanner
	logger            logging.Logger
}

// NewPeerManager creates a new peer manager. The provided context is used as the base context for RPC connections
// established by the manager.
func NewPeerManager(
	ctx context.Context,
	config PeerManagerConfig,
	dialer Dialer,
	messageReplicator MessageReplicator,
	blobReplicator BlobReplicator,
	roomScanner RoomScanner,
	logger logging.Logger,
) *PeerManager {
	return &PeerManager{
		ctx:               ctx,
		peers:             make(peersMap),
		peersLock:         &sync.Mutex{},
		config:            config,
		dialer:            dialer,
		messageReplicator: messageReplicator,
		blobReplicator:    blobReplicator,
		roomScanner:       roomScanner,
		logger:            logger.New("peer_manager"),
	}
}

// EstablishNewConnections tries to establish new connections to the preferred pubs.
func (p PeerManager) EstablishNewConnections() error {
	var resultErr error

	for _, pub := range p.config.PreferredPubs {
		if err := p.Connect(pub.Identity, pub.Address); err != nil {
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

// Connect attempts to establish communications with the specified peer. If a connection to the specified peer
// already exists then a new connection will not be initiated. If connecting to the peer succeeds but in the meantime
// a connection to the same node was created manually or automatically by the manager then the old connection will be
// replaced by the new connection and terminated.
func (p PeerManager) Connect(remote identity.Public, address network.Address) error {
	select {
	case <-p.ctx.Done():
		return errors.Wrap(p.ctx.Err(), "context is done so the connection would just terminate right away")
	default:
	}

	if p.alreadyConnected(remote) { // early check
		return nil
	}

	p.logger.WithField("remote", remote).WithField("address", address).Debug("dialing")
	defer p.logger.Debug("dial exiting")

	peer, err := p.dialer.Dial(p.ctx, remote, address)
	if err != nil {
		return errors.Wrap(err, "dial failed")
	}

	p.HandleNewPeer(peer)

	return nil
}

// ProcessNewLocalDiscovery handles incoming local peer announcements.
func (p PeerManager) ProcessNewLocalDiscovery(remote identity.Public, address network.Address) error {
	return p.Connect(remote, address)
}

// HandleNewPeer registers a new peer with the connection manager. It should be called for every new peer e.g. when
// an incoming connection is received. Do not call this method if a connection was initialized by the manager e.g.
// through calling Connect or EstablishNewConnections.
func (p PeerManager) HandleNewPeer(peer transport.Peer) {
	p.trackPeer(peer)
	go p.processConnection(peer)
}

// trackPeer adds a peer to the list of peers. If we are already communicating with this identity the previous
// connection to it is terminated.
func (p PeerManager) trackPeer(peer transport.Peer) {
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
	go p.removePeerIfConnectionCloses(peer)
}

func (p PeerManager) removePeerIfConnectionCloses(peer transport.Peer) {
	<-peer.Conn().Context().Done()

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

// todo this probably shouldn't be handled by the peer manager
func (p PeerManager) processConnection(peer transport.Peer) {
	p.logger.WithField("peer", peer).Debug("handling a new peer")
	if err := p.runTasks(peer); err != nil {
		p.logger.WithError(err).WithField("peer", peer).Debug("all tasks ended")
	}
}

func (p PeerManager) runTasks(peer transport.Peer) error {
	ch := make(chan error)

	ctx, cancel := context.WithCancel(peer.Conn().Context())
	defer cancel()

	tasks := 0

	p.startTask(&tasks, ctx, peer, ch, p.messageReplicator.Replicate, "message replication")
	p.startTask(&tasks, ctx, peer, ch, p.blobReplicator.Replicate, "blob replication")
	p.startTask(&tasks, ctx, peer, ch, p.roomScanner.Run, "room scanner")

	var result error
	for i := 0; i < tasks; i++ {
		result = multierror.Append(result, <-ch)
	}
	return result
}

func (p PeerManager) startTask(
	tasks *int,
	ctx context.Context,
	peer transport.Peer,
	ch chan<- error,
	fn func(ctx context.Context, peer transport.Peer) error,
	taskName string,
) {
	peerLogger := p.logger.WithField("peer", peer)
	*tasks = *tasks + 1
	go func() {
		err := fn(ctx, peer)
		peerLogger.WithError(err).WithField("task", taskName).Debug("task terminating")
		ch <- err
	}()

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
