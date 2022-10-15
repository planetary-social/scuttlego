package domain_test

import (
	"bytes"
	"context"
	"sort"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestPeerManager_PeersAreTracked(t *testing.T) {
	m := buildTestPeerManager(t)

	peer1 := transport.NewPeer(fixtures.SomePublicIdentity(), newConnectionMock())
	m.Manager.HandleNewPeer(peer1)

	require.Equal(t,
		[]transport.Peer{
			peer1,
		},
		m.Manager.Peers(),
	)

	peer2 := transport.NewPeer(fixtures.SomePublicIdentity(), newConnectionMock())
	m.Manager.HandleNewPeer(peer2)

	expectedPeers := []transport.Peer{
		peer1,
		peer2,
	}

	actualPeers := m.Manager.Peers()

	sortPeers(expectedPeers)
	sortPeers(actualPeers)

	require.Equal(t,
		expectedPeers,
		actualPeers,
	)
}

func TestPeerManager_IfConnectionClosesThenPeerIsRemovedFromManager(t *testing.T) {
	m := buildTestPeerManager(t)

	conn := newConnectionMock()
	peer := transport.NewPeer(fixtures.SomePublicIdentity(), conn)
	m.Manager.HandleNewPeer(peer)

	require.Equal(t,
		[]transport.Peer{
			peer,
		},
		m.Manager.Peers(),
	)

	err := conn.Close()
	require.NoError(t, err)

	eventually(t,
		func() bool {
			return len(m.Manager.Peers()) == 0
		},
		"if a connection is closed then the manager should eventually remove a peer from its list",
	)
}

func TestPeerManager_OnlyLatestConnectionToIdentityIsKept(t *testing.T) {
	m := buildTestPeerManager(t)

	iden := fixtures.SomePublicIdentity()

	conn1 := newConnectionMock()
	peer1 := transport.NewPeer(iden, conn1)
	m.Manager.HandleNewPeer(peer1)

	require.Equal(t,
		[]transport.Peer{
			peer1,
		},
		m.Manager.Peers(),
	)

	conn2 := newConnectionMock()
	peer2 := transport.NewPeer(iden, conn2)
	m.Manager.HandleNewPeer(peer2)

	require.Equal(t,
		[]transport.Peer{
			peer2,
		},
		m.Manager.Peers(),
	)

	eventually(t,
		func() bool {
			return conn1.IsClosed()
		},
		"if a peer is dropped by the manager then the manager should close its connection",
	)
}

func TestPeerManager_Connect_DialsAnIdentityIfNotConnectedToIt(t *testing.T) {
	m := buildTestPeerManager(t)

	address := network.NewAddress("some address")
	peer := transport.NewPeer(fixtures.SomePublicIdentity(), newConnectionMock())
	m.Dialer.AddPeer(peer, address)

	require.Empty(t, m.Manager.Peers())

	err := m.Manager.Connect(peer.Identity(), address)
	require.NoError(t, err)

	require.Equal(t,
		[]identity.Public{
			peer.Identity(),
		},
		m.Dialer.DialedPeers,
	)

	require.Equal(t,
		[]transport.Peer{
			peer,
		},
		m.Manager.Peers(),
	)
}

func TestPeerManager_Connect_DoesNotDialIfAlreadyConnectedToIdentity(t *testing.T) {
	m := buildTestPeerManager(t)

	iden := fixtures.SomePublicIdentity()
	alreadyConnectedPeer := transport.NewPeer(iden, newConnectionMock())
	m.Manager.HandleNewPeer(alreadyConnectedPeer)

	err := m.Manager.Connect(iden, network.NewAddress("someAddress"))
	require.NoError(t, err)

	require.Empty(t, m.Dialer.DialedPeers)
}

func TestPeerManager_EstablishNewConnections_ConnectsToPreferredPubs(t *testing.T) {
	pub := domain.Pub{
		Identity: fixtures.SomePublicIdentity(),
		Address:  network.NewAddress("some address"),
	}

	config := domain.PeerManagerConfig{
		PreferredPubs: []domain.Pub{
			pub,
		},
	}

	m := buildTestPeerManagerWithConfig(t, config)

	peer := transport.NewPeer(pub.Identity, newConnectionMock())
	m.Dialer.AddPeer(peer, pub.Address)

	require.Empty(t, m.Manager.Peers())

	err := m.Manager.EstablishNewConnections()
	require.NoError(t, err)

	require.Equal(t,
		[]identity.Public{
			peer.Identity(),
		},
		m.Dialer.DialedPeers,
	)

	require.Equal(t,
		[]transport.Peer{
			peer,
		},
		m.Manager.Peers(),
	)
}

func TestPeerManager_EstablishNewConnections_DoesNotConnectToPreferredPubsIfAlreadyConnected(t *testing.T) {
	pub := domain.Pub{
		Identity: fixtures.SomePublicIdentity(),
		Address:  network.NewAddress("some address"),
	}

	config := domain.PeerManagerConfig{
		PreferredPubs: []domain.Pub{
			pub,
		},
	}

	m := buildTestPeerManagerWithConfig(t, config)

	peer := transport.NewPeer(pub.Identity, newConnectionMock())
	m.Dialer.AddPeer(peer, pub.Address)

	err := m.Manager.EstablishNewConnections()
	require.NoError(t, err)

	err = m.Manager.EstablishNewConnections()
	require.NoError(t, err)

	require.Len(t, m.Dialer.DialedPeers, 1)
}

func TestPeerManager_ProcessNewLocalDiscovery_ConnectsToPeerOnDiscovery(t *testing.T) {
	m := buildTestPeerManager(t)

	address := network.NewAddress("some address")
	peer := transport.NewPeer(fixtures.SomePublicIdentity(), newConnectionMock())
	m.Dialer.AddPeer(peer, address)

	require.Empty(t, m.Manager.Peers())

	err := m.Manager.ProcessNewLocalDiscovery(peer.Identity(), address)
	require.NoError(t, err)

	require.Equal(t,
		[]identity.Public{
			peer.Identity(),
		},
		m.Dialer.DialedPeers,
	)

	require.Equal(t,
		[]transport.Peer{
			peer,
		},
		m.Manager.Peers(),
	)
}

func TestPeerManager_ProcessNewLocalDiscovery_DoesNotConnectIfAlreadyConnected(t *testing.T) {
	m := buildTestPeerManager(t)

	iden := fixtures.SomePublicIdentity()
	alreadyConnectedPeer := transport.NewPeer(iden, newConnectionMock())
	m.Manager.HandleNewPeer(alreadyConnectedPeer)

	err := m.Manager.ProcessNewLocalDiscovery(iden, network.NewAddress("someAddress"))
	require.NoError(t, err)

	require.Empty(t, m.Dialer.DialedPeers)
}

type testPeerManager struct {
	Manager *domain.PeerManager
	Dialer  *dialerMock
}

func buildTestPeerManagerWithConfig(t *testing.T, config domain.PeerManagerConfig) testPeerManager {
	ctx := fixtures.TestContext(t)
	logger := logging.NewDevNullLogger()

	msgReplicator := newMessageReplicatorMock()
	blobReplicator := newBlobReplicatorMock()
	dialer := newDialerMock()

	manager := domain.NewPeerManager(ctx, config, msgReplicator, blobReplicator, dialer, logger)

	return testPeerManager{
		Manager: manager,
		Dialer:  dialer,
	}
}

func buildTestPeerManager(t *testing.T) testPeerManager {
	return buildTestPeerManagerWithConfig(t, domain.PeerManagerConfig{})
}

type messageReplicatorMock struct {
}

func newMessageReplicatorMock() *messageReplicatorMock {
	return &messageReplicatorMock{}
}

func (r messageReplicatorMock) Replicate(ctx context.Context, peer transport.Peer) error {
	<-ctx.Done()
	return ctx.Err()
}

type blobReplicatorMock struct {
}

func newBlobReplicatorMock() *blobReplicatorMock {
	return &blobReplicatorMock{}
}

func (b blobReplicatorMock) Replicate(ctx context.Context, peer transport.Peer) error {
	<-ctx.Done()
	return ctx.Err()
}

type dialerMock struct {
	peers       map[network.Address]transport.Peer
	DialedPeers []identity.Public
}

func newDialerMock() *dialerMock {
	return &dialerMock{
		peers: make(map[network.Address]transport.Peer),
	}
}

func (d *dialerMock) AddPeer(peer transport.Peer, address network.Address) {
	d.peers[address] = peer
}

func (d *dialerMock) Dial(ctx context.Context, remote identity.Public, address network.Address) (transport.Peer, error) {
	d.DialedPeers = append(d.DialedPeers, remote)
	peer, ok := d.peers[address]
	if !ok {
		return transport.Peer{}, errors.New("peer not found")
	}

	if !peer.Identity().PublicKey().Equal(remote.PublicKey()) {
		return transport.Peer{}, errors.New("unexpected peer identity")
	}

	return peer, nil
}

type connectionMock struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func newConnectionMock() *connectionMock {
	ctx, cancel := context.WithCancel(context.Background())
	return &connectionMock{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c connectionMock) WasInitiatedByRemote() bool {
	return fixtures.SomeBool()
}

func (c connectionMock) PerformRequest(ctx context.Context, req *rpc.Request) (rpc.ResponseStream, error) {
	return nil, errors.New("not implemented")
}

func (c connectionMock) Context() context.Context {
	return c.ctx
}

func (c *connectionMock) Close() error {
	c.cancel()
	return nil
}

func (c *connectionMock) IsClosed() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}

func eventually(t *testing.T, condition func() bool, msgAndArgs ...any) {
	require.Eventually(t, condition, 1*time.Second, 10*time.Millisecond, msgAndArgs)
}

func sortPeers(peers []transport.Peer) {
	sort.Slice(peers, func(i, j int) bool {
		return bytes.Compare(peers[i].Identity().PublicKey(), peers[j].Identity().PublicKey()) < 0
	})
}
