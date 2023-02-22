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
	ctx := fixtures.TestContext(t)

	peer1 := transport.MustNewPeer(fixtures.SomePublicIdentity(), newConnectionMock())
	m.Manager.TrackPeer(ctx, peer1)

	require.Equal(t,
		[]transport.Peer{
			peer1,
		},
		m.Manager.Peers(),
	)

	peer2 := transport.MustNewPeer(fixtures.SomePublicIdentity(), newConnectionMock())
	m.Manager.TrackPeer(ctx, peer2)

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

func TestPeerManager_IfContextClosesThenPeerIsRemovedFromManager(t *testing.T) {
	m := buildTestPeerManager(t)
	ctx := fixtures.TestContext(t)

	conn := newConnectionMock()
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), conn)

	ctx, cancel := context.WithCancel(ctx)
	m.Manager.TrackPeer(ctx, peer)

	require.Equal(t,
		[]transport.Peer{
			peer,
		},
		m.Manager.Peers(),
	)

	cancel()

	eventually(t,
		func() bool {
			return len(m.Manager.Peers()) == 0
		},
		"if a connection is closed then the manager should eventually remove a peer from its list",
	)
}

func TestPeerManager_OnlyLatestConnectionToIdentityIsKept(t *testing.T) {
	m := buildTestPeerManager(t)
	ctx := fixtures.TestContext(t)

	iden := fixtures.SomePublicIdentity()

	conn1 := newConnectionMock()
	peer1 := transport.MustNewPeer(iden, conn1)
	m.Manager.TrackPeer(ctx, peer1)

	require.Equal(t,
		[]transport.Peer{
			peer1,
		},
		m.Manager.Peers(),
	)

	conn2 := newConnectionMock()
	peer2 := transport.MustNewPeer(iden, conn2)
	m.Manager.TrackPeer(ctx, peer2)

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
	ctx := fixtures.TestContext(t)

	address := network.NewAddress("some address")
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), newConnectionMock())
	m.Dialer.AddPeer(peer, address)

	require.Empty(t, m.Manager.Peers())

	err := m.Manager.Connect(ctx, peer.Identity(), address)
	require.NoError(t, err)

	require.Equal(t,
		[]identity.Public{
			peer.Identity(),
		},
		m.Dialer.DialedPeers,
	)
}

func TestPeerManager_Connect_DoesNotDialIfAlreadyConnectedToIdentity(t *testing.T) {
	m := buildTestPeerManager(t)
	ctx := fixtures.TestContext(t)

	iden := fixtures.SomePublicIdentity()
	alreadyConnectedPeer := transport.MustNewPeer(iden, newConnectionMock())
	m.Manager.TrackPeer(ctx, alreadyConnectedPeer)

	err := m.Manager.Connect(ctx, iden, network.NewAddress("someAddress"))
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
	ctx := fixtures.TestContext(t)

	peer := transport.MustNewPeer(pub.Identity, newConnectionMock())
	m.Dialer.AddPeer(peer, pub.Address)

	require.Empty(t, m.Manager.Peers())

	err := m.Manager.EstablishNewConnections(ctx)
	require.NoError(t, err)

	require.Equal(t,
		[]identity.Public{
			peer.Identity(),
		},
		m.Dialer.DialedPeers,
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
	ctx := fixtures.TestContext(t)

	peer := transport.MustNewPeer(pub.Identity, newConnectionMock())
	m.Dialer.AddPeer(peer, pub.Address)

	m.Manager.TrackPeer(ctx, peer)

	err := m.Manager.EstablishNewConnections(ctx)
	require.NoError(t, err)

	require.Len(t, m.Dialer.DialedPeers, 0)
}

func TestPeerManager_ProcessNewLocalDiscovery_ConnectsToPeerOnDiscovery(t *testing.T) {
	m := buildTestPeerManager(t)
	ctx := fixtures.TestContext(t)

	address := network.NewAddress("some address")
	peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), newConnectionMock())
	m.Dialer.AddPeer(peer, address)

	require.Empty(t, m.Manager.Peers())

	err := m.Manager.ProcessNewLocalDiscovery(ctx, peer.Identity(), address)
	require.NoError(t, err)

	require.Equal(t,
		[]identity.Public{
			peer.Identity(),
		},
		m.Dialer.DialedPeers,
	)
}

func TestPeerManager_ProcessNewLocalDiscovery_DoesNotConnectIfAlreadyConnected(t *testing.T) {
	m := buildTestPeerManager(t)
	ctx := fixtures.TestContext(t)

	iden := fixtures.SomePublicIdentity()
	alreadyConnectedPeer := transport.MustNewPeer(iden, newConnectionMock())
	m.Manager.TrackPeer(ctx, alreadyConnectedPeer)

	err := m.Manager.ProcessNewLocalDiscovery(ctx, iden, network.NewAddress("someAddress"))
	require.NoError(t, err)

	require.Empty(t, m.Dialer.DialedPeers)
}

type testPeerManager struct {
	Manager *domain.PeerManager
	Dialer  *dialerMock
}

func buildTestPeerManagerWithConfig(t *testing.T, config domain.PeerManagerConfig) testPeerManager {
	logger := logging.NewDevNullLogger()

	dialer := newDialerMock()
	roomDialer := newRoomDialerMock()

	manager := domain.NewPeerManager(config, dialer, roomDialer, logger)

	return testPeerManager{
		Manager: manager,
		Dialer:  dialer,
	}
}

func buildTestPeerManager(t *testing.T) testPeerManager {
	return buildTestPeerManagerWithConfig(t, domain.PeerManagerConfig{})
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

type roomDialerMock struct {
}

func newRoomDialerMock() *roomDialerMock {
	return &roomDialerMock{}
}

func (r roomDialerMock) DialViaRoom(ctx context.Context, portal transport.Peer, target identity.Public) (transport.Peer, error) {
	return transport.Peer{}, errors.New("not implemented")
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
