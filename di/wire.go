//go:build wireinject
// +build wireinject

package di

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/adapters/bolt"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/adapters/pubsub"
	"github.com/planetary-social/scuttlego/service/app"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain"
	blobReplication "github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	domainmocks "github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/network/local"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/planetary-social/scuttlego/service/domain/replication/gossip"
	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/rooms/tunnel"
	"go.etcd.io/bbolt"
)

type TxTestAdapters struct {
	MessageRepository     *bolt.MessageRepository
	FeedRepository        *bolt.FeedRepository
	BlobRepository        *bolt.BlobRepository
	SocialGraphRepository *bolt.SocialGraphRepository
	PubRepository         *bolt.PubRepository
	ReceiveLog            *bolt.ReceiveLogRepository
	BlobWantList          *bolt.BlobWantListRepository
	FeedWantList          *bolt.FeedWantListRepository
	BanList               *bolt.BanListRepository

	CurrentTimeProvider *mocks.CurrentTimeProviderMock
	BanListHasher       *mocks.BanListHasherMock
}

func BuildTxTestAdapters(*bbolt.Tx) (TxTestAdapters, error) {
	wire.Build(
		wire.Struct(new(TxTestAdapters), "*"),

		txBoltAdaptersSet,
		testAdaptersSet,

		identity.NewPrivate,
		privateIdentityToPublicIdentity,

		formats.NewDefaultMessageHMAC,

		fixtures.SomeLogger,

		formatsSet,
		wire.Value(hops),
	)

	return TxTestAdapters{}, nil
}

type TestAdapters struct {
	MessageRepository *bolt.ReadMessageRepository
	FeedRepository    *bolt.ReadFeedRepository
	ReceiveLog        *bolt.ReadReceiveLogRepository
}

func BuildTestAdapters(*bbolt.DB) (TestAdapters, error) {
	wire.Build(
		wire.Struct(new(TestAdapters), "*"),

		boltAdaptersSet,

		identity.NewPrivate,
		privateIdentityToPublicIdentity,

		formats.NewDefaultMessageHMAC,

		fixtures.SomeLogger,
	)

	return TestAdapters{}, nil
}

type TestCommands struct {
	RoomsAliasRegister        *commands.RoomsAliasRegisterHandler
	RoomsAliasRevoke          *commands.RoomsAliasRevokeHandler
	ProcessRoomAttendantEvent *commands.ProcessRoomAttendantEventHandler
	DisconnectAll             *commands.DisconnectAllHandler
	DownloadFeed              *commands.DownloadFeedHandler

	PeerManager            *domainmocks.PeerManagerMock
	Dialer                 *mocks.DialerMock
	FeedWantListRepository *mocks.FeedWantListRepositoryMock
	CurrentTimeProvider    *mocks.CurrentTimeProviderMock
}

func BuildTestCommands(*testing.T) (TestCommands, error) {
	wire.Build(
		applicationSet,

		mocks.NewDialerMock,
		wire.Bind(new(commands.Dialer), new(*mocks.DialerMock)),

		domainmocks.NewPeerManagerMock,
		wire.Bind(new(commands.PeerManager), new(*domainmocks.PeerManagerMock)),

		identity.NewPrivate,

		mocks.NewMockTransactionProvider,
		wire.Bind(new(commands.TransactionProvider), new(*mocks.MockTransactionProvider)),

		wire.Struct(new(commands.Adapters), "FeedWantList"),

		mocks.NewFeedWantListRepositoryMock,
		wire.Bind(new(commands.FeedWantListRepository), new(*mocks.FeedWantListRepositoryMock)),

		mocks.NewCurrentTimeProviderMock,
		wire.Bind(new(commands.CurrentTimeProvider), new(*mocks.CurrentTimeProviderMock)),

		wire.Struct(new(TestCommands), "*"),
	)

	return TestCommands{}, nil
}

type TestQueries struct {
	Queries app.Queries

	FeedRepository       *mocks.FeedRepositoryMock
	MessagePubSub        *mocks.MessagePubSubMock
	MessageRepository    *mocks.MessageRepositoryMock
	PeerManager          *domainmocks.PeerManagerMock
	BlobStorage          *mocks.BlobStorageMock
	ReceiveLogRepository *mocks.ReceiveLogRepositoryMock
	Dialer               *mocks.DialerMock

	LocalIdentity identity.Public
}

func BuildTestQueries(*testing.T) (TestQueries, error) {
	wire.Build(
		applicationSet,
		mockQueryAdaptersSet,

		pubsub.NewMessagePubSub,
		mocks.NewMessagePubSubMock,
		wire.Bind(new(queries.MessageSubscriber), new(*mocks.MessagePubSubMock)),

		domainmocks.NewPeerManagerMock,
		wire.Bind(new(queries.PeerManager), new(*domainmocks.PeerManagerMock)),

		identity.NewPrivate,
		privateIdentityToPublicIdentity,

		mocks.NewBlobStorageMock,
		wire.Bind(new(queries.BlobStorage), new(*mocks.BlobStorageMock)),

		mocks.NewBlobDownloadedPubSubMock,
		wire.Bind(new(queries.BlobDownloadedSubscriber), new(*mocks.BlobDownloadedPubSubMock)),

		mocks.NewDialerMock,
		wire.Bind(new(queries.Dialer), new(*mocks.DialerMock)),

		wire.Struct(new(TestQueries), "*"),

		fixtures.TestLogger,
	)

	return TestQueries{}, nil
}

func BuildTransactableAdapters(*bbolt.Tx, identity.Public, Config) (commands.Adapters, error) {
	wire.Build(
		wire.Struct(new(commands.Adapters), "*"),

		txBoltAdaptersSet,
		formatsSet,
		extractFromConfigSet,
		adaptersSet,

		wire.Value(hops),
	)

	return commands.Adapters{}, nil
}

func BuildTxRepositories(*bbolt.Tx, identity.Public, logging.Logger, formats.MessageHMAC) (bolt.TxRepositories, error) {
	wire.Build(
		wire.Struct(new(bolt.TxRepositories), "*"),

		txBoltAdaptersSet,
		formatsSet,
		adaptersSet,

		wire.Value(hops),
	)

	return bolt.TxRepositories{}, nil
}

// BuildService creates a new service which uses the provided context as a long-term context used as a base context for
// e.g. established connections.
func BuildService(context.Context, identity.Private, Config) (Service, error) {
	wire.Build(
		NewService,

		domain.NewPeerManager,
		wire.Bind(new(commands.NewPeerHandler), new(*domain.PeerManager)),
		wire.Bind(new(commands.PeerManager), new(*domain.PeerManager)),
		wire.Bind(new(queries.PeerManager), new(*domain.PeerManager)),

		bolt.NewTransactionProvider,
		wire.Bind(new(commands.TransactionProvider), new(*bolt.TransactionProvider)),
		newAdaptersFactory,

		newBolt,

		newAdvertiser,
		privateIdentityToPublicIdentity,

		commands.NewMessageBuffer,

		rooms.NewScanner,
		wire.Bind(new(domain.RoomScanner), new(*rooms.Scanner)),

		rooms.NewPeerRPCAdapter,
		wire.Bind(new(rooms.MetadataGetter), new(*rooms.PeerRPCAdapter)),
		wire.Bind(new(rooms.AttendantsGetter), new(*rooms.PeerRPCAdapter)),

		tunnel.NewDialer,
		wire.Bind(new(domain.RoomDialer), new(*tunnel.Dialer)),

		portsSet,
		applicationSet,
		replicatorSet,
		blobReplicatorSet,
		formatsSet,
		pubSubSet,
		boltAdaptersSet,
		blobsAdaptersSet,
		adaptersSet,
		extractFromConfigSet,
		networkingSet,
	)
	return Service{}, nil
}

var replicatorSet = wire.NewSet(
	gossip.NewManager,
	wire.Bind(new(gossip.ReplicationManager), new(*gossip.Manager)),

	gossip.NewGossipReplicator,
	wire.Bind(new(replication.CreateHistoryStreamReplicator), new(*gossip.GossipReplicator)),

	ebt.NewReplicator,
	wire.Bind(new(replication.EpidemicBroadcastTreesReplicator), new(ebt.Replicator)),

	replication.NewContactsCache,
	wire.Bind(new(gossip.ContactsStorage), new(*replication.ContactsCache)),
	wire.Bind(new(ebt.ContactsStorage), new(*replication.ContactsCache)),

	ebt.NewSessionTracker,
	wire.Bind(new(ebt.Tracker), new(*ebt.SessionTracker)),

	ebt.NewSessionRunner,
	wire.Bind(new(ebt.Runner), new(*ebt.SessionRunner)),

	replication.NewNegotiator,
	wire.Bind(new(domain.MessageReplicator), new(*replication.Negotiator)),
)

var blobReplicatorSet = wire.NewSet(
	blobReplication.NewManager,
	wire.Bind(new(blobReplication.ReplicationManager), new(*blobReplication.Manager)),
	wire.Bind(new(commands.BlobReplicationManager), new(*blobReplication.Manager)),

	blobReplication.NewReplicator,
	wire.Bind(new(domain.BlobReplicator), new(*blobReplication.Replicator)),

	blobReplication.NewBlobsGetDownloader,
	wire.Bind(new(blobReplication.Downloader), new(*blobReplication.BlobsGetDownloader)),

	blobReplication.NewHasHandler,
	wire.Bind(new(blobReplication.HasBlobHandler), new(*blobReplication.HasHandler)),
)

var hops = graph.MustNewHops(3)

func newAdvertiser(l identity.Public, config Config) (*local.Advertiser, error) {
	return local.NewAdvertiser(l, config.ListenAddress)
}

func newAdaptersFactory(config Config, local identity.Public) bolt.AdaptersFactory {
	return func(tx *bbolt.Tx) (commands.Adapters, error) {
		return BuildTransactableAdapters(tx, local, config)
	}
}

func newBolt(config Config) (*bbolt.DB, error) {
	filename := path.Join(config.DataDirectory, "database.bolt")
	b, err := bbolt.Open(filename, 0600, &bbolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, errors.Wrap(err, "could not open the database, is something else reading it?")
	}
	return b, nil
}

func privateIdentityToPublicIdentity(p identity.Private) identity.Public {
	return p.Public()
}
