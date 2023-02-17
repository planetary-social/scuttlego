//go:build wireinject
// +build wireinject

package di

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/logging"
	badgeradapters "github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/adapters/badger/notx"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/adapters/pubsub"
	"github.com/planetary-social/scuttlego/service/app"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/transport"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/invites"
	domainmocks "github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/network/local"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/planetary-social/scuttlego/service/domain/replication/gossip"
	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/rooms/tunnel"
)

type BadgerNoTxTestAdapters struct {
	NoTxTestAdapters    notx.TestAdapters
	TransactionProvider *badgeradapters.TestTransactionProvider
	Dependencies        *badgeradapters.TestAdaptersDependencies
}

func BuildBadgerNoTxTestAdapters(t *testing.T) BadgerNoTxTestAdapters {
	wire.Build(
		wire.Struct(new(BadgerNoTxTestAdapters), "*"),

		wire.Struct(new(notx.TestAdapters), "*"),

		badgerNoTxTestTransactionProviderSet,
		badgerNoTxRepositoriesSet,
		testBadgerTransactionProviderSet,
		badgerTestAdaptersDependenciesSet,

		fixtures.SomePublicIdentity,
		fixtures.Badger,
	)

	return BadgerNoTxTestAdapters{}
}

type BadgerTestAdapters struct {
	TransactionProvider *badgeradapters.TestTransactionProvider
	Dependencies        *badgeradapters.TestAdaptersDependencies
}

func BuildBadgerTestAdapters(t *testing.T) BadgerTestAdapters {
	wire.Build(
		wire.Struct(new(BadgerTestAdapters), "*"),

		testBadgerTransactionProviderSet,
		badgerTestAdaptersDependenciesSet,

		fixtures.SomePublicIdentity,
		fixtures.Badger,
	)

	return BadgerTestAdapters{}
}

func buildTestBadgerNoTxTxAdapters(*badger.Txn, badgeradapters.TestAdaptersDependencies) (notx.TxAdapters, error) {
	wire.Build(
		wire.Struct(new(notx.TxAdapters), "*"),

		badgerRepositoriesSet,
		badgerUnpackTestDependenciesSet,
		contentSet,

		formats.NewDefaultMessageHMAC,
		formats.NewScuttlebutt,
		transport.DefaultMappings,

		transport.NewMarshaler,
		wire.Bind(new(content.Marshaler), new(*transport.Marshaler)),

		fixtures.SomeLogger,
		fixtures.SomeHops,
	)

	return notx.TxAdapters{}, nil
}

func buildBadgerNoTxTxAdapters(*badger.Txn, identity.Public, Config, logging.Logger) (notx.TxAdapters, error) {
	wire.Build(
		wire.Struct(new(notx.TxAdapters), "*"),

		badgerRepositoriesSet,
		formatsSet,
		extractFromConfigSet,
		adaptersSet,
		contentSet,
	)

	return notx.TxAdapters{}, nil
}

func buildBadgerTestAdapters(*badger.Txn, badgeradapters.TestAdaptersDependencies) (badgeradapters.TestAdapters, error) {
	wire.Build(
		wire.Struct(new(badgeradapters.TestAdapters), "*"),

		badgerRepositoriesSet,
		badgerUnpackTestDependenciesSet,
		contentSet,

		formats.NewDefaultMessageHMAC,
		formats.NewScuttlebutt,
		transport.DefaultMappings,

		transport.NewMarshaler,
		wire.Bind(new(content.Marshaler), new(*transport.Marshaler)),

		fixtures.SomeLogger,
		fixtures.SomeHops,
	)

	return badgeradapters.TestAdapters{}, nil
}

type TestCommands struct {
	RoomsAliasRegister        *commands.RoomsAliasRegisterHandler
	RoomsAliasRevoke          *commands.RoomsAliasRevokeHandler
	ProcessRoomAttendantEvent *commands.ProcessRoomAttendantEventHandler
	DisconnectAll             *commands.DisconnectAllHandler
	DownloadFeed              *commands.DownloadFeedHandler
	RedeemInvite              *commands.RedeemInviteHandler
	AcceptTunnelConnect       *commands.AcceptTunnelConnectHandler

	MigrationImportDataFromGoSSB *commands.MigrationHandlerImportDataFromGoSSB

	PeerManager            *domainmocks.PeerManagerMock
	Dialer                 *mocks.DialerMock
	FeedWantListRepository *mocks.FeedWantListRepositoryMock
	CurrentTimeProvider    *mocks.CurrentTimeProviderMock
	InviteRedeemer         *mocks.InviteRedeemerMock
	Local                  identity.Public
	PeerInitializer        *mocks.PeerInitializerMock
	NewPeerHandler         *mocks.NewPeerHandlerMock
	GoSSBRepoReader        *mocks.GoSSBRepoReaderMock
	FeedRepository         *mocks.FeedRepositoryMock
	ReceiveLog             *mocks.ReceiveLogRepositoryMock
}

func BuildTestCommands(*testing.T) (TestCommands, error) {
	wire.Build(
		commandsSet,
		migrationCommandsSet,

		mocks.NewDialerMock,
		wire.Bind(new(commands.Dialer), new(*mocks.DialerMock)),

		domainmocks.NewPeerManagerMock,
		wire.Bind(new(commands.PeerManager), new(*domainmocks.PeerManagerMock)),

		identity.NewPrivate,
		privateIdentityToPublicIdentity,

		mocks.NewMockCommandsTransactionProvider,
		wire.Bind(new(commands.TransactionProvider), new(*mocks.MockCommandsTransactionProvider)),

		wire.Struct(
			new(commands.Adapters),
			"FeedWantList",
			"Feed",
			"ReceiveLog",
		),

		mocks.NewFeedWantListRepositoryMock,
		wire.Bind(new(commands.FeedWantListRepository), new(*mocks.FeedWantListRepositoryMock)),

		mocks.NewCurrentTimeProviderMock,
		wire.Bind(new(commands.CurrentTimeProvider), new(*mocks.CurrentTimeProviderMock)),

		mocks.NewInviteRedeemerMock,
		wire.Bind(new(commands.InviteRedeemer), new(*mocks.InviteRedeemerMock)),

		mocks.NewPeerInitializerMock,
		wire.Bind(new(commands.ServerPeerInitializer), new(*mocks.PeerInitializerMock)),

		mocks.NewNewPeerHandlerMock,
		wire.Bind(new(commands.NewPeerHandler), new(*mocks.NewPeerHandlerMock)),

		mocks.NewGoSSBRepoReaderMock,
		wire.Bind(new(commands.GoSSBRepoReader), new(*mocks.GoSSBRepoReaderMock)),

		mocks.NewContentParser,
		wire.Bind(new(commands.ContentParser), new(*mocks.ContentParser)),

		mocks.NewFeedRepositoryMock,
		wire.Bind(new(commands.FeedRepository), new(*mocks.FeedRepositoryMock)),

		mocks.NewReceiveLogRepositoryMock,
		wire.Bind(new(commands.ReceiveLogRepository), new(*mocks.ReceiveLogRepositoryMock)),

		fixtures.TestLogger,

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

		mocks.NewMockQueriesTransactionProvider,
		wire.Bind(new(queries.TransactionProvider), new(*mocks.MockQueriesTransactionProvider)),

		wire.Struct(
			new(queries.Adapters),
			"Feed",
			"ReceiveLog",
			"Message",
		),

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

func buildBadgerCommandsAdapters(*badger.Txn, identity.Public, Config, logging.Logger) (commands.Adapters, error) {
	wire.Build(
		wire.Struct(new(commands.Adapters), "*"),

		badgerRepositoriesSet,
		formatsSet,
		extractFromConfigSet,
		adaptersSet,
		contentSet,
	)

	return commands.Adapters{}, nil
}

func buildBadgerQueriesAdapters(*badger.Txn, identity.Public, Config, logging.Logger) (queries.Adapters, error) {
	wire.Build(
		wire.Struct(new(queries.Adapters), "*"),

		badgerRepositoriesSet,
		formatsSet,
		extractFromConfigSet,
		adaptersSet,
		contentSet,
	)

	return queries.Adapters{}, nil
}

// BuildService creates a new service which uses the provided context as a long-term context used as a base context for
// e.g. established connections.
func BuildService(context.Context, identity.Private, Config) (Service, func(), error) {
	wire.Build(
		NewService,

		domain.NewPeerManager,
		wire.Bind(new(commands.NewPeerHandler), new(*domain.PeerManager)),
		wire.Bind(new(commands.PeerManager), new(*domain.PeerManager)),
		wire.Bind(new(queries.PeerManager), new(*domain.PeerManager)),

		newBadger,

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

		invites.NewInviteRedeemer,
		wire.Bind(new(commands.InviteRedeemer), new(*invites.InviteRedeemer)),

		commands.NewTransactionRawMessagePublisher,
		wire.Bind(new(commands.RawMessagePublisher), new(*commands.TransactionRawMessagePublisher)),

		newContextLogger,

		portsSet,
		applicationSet,
		replicatorSet,
		blobReplicatorSet,
		formatsSet,
		pubSubSet,
		badgerNoTxRepositoriesSet,
		badgerTransactionProviderSet,
		badgerNoTxTransactionProviderSet,
		badgerAdaptersSet,
		blobsAdaptersSet,
		adaptersSet,
		extractFromConfigSet,
		networkingSet,
		migrationsSet,
		contentSet,
	)
	return Service{}, nil, nil
}

type IntegrationTestsService struct {
	Service Service

	BanListHasher badgeradapters.BanListHasher
}

func BuildIntegrationTestsService(t *testing.T) (IntegrationTestsService, error) {
	service, cleanup, err := buildIntegrationTestsService(t)
	if err != nil {
		return IntegrationTestsService{}, errors.Wrap(err, "error calling wire builder")
	}
	t.Cleanup(cleanup)
	return service, nil
}

func buildIntegrationTestsService(t *testing.T) (IntegrationTestsService, func(), error) {
	wire.Build(
		wire.Struct(new(IntegrationTestsService), "*"),

		newIntegrationTestConfig,
		fixtures.SomePrivateIdentity,
		fixtures.TestContext,

		NewService,

		domain.NewPeerManager,
		wire.Bind(new(commands.NewPeerHandler), new(*domain.PeerManager)),
		wire.Bind(new(commands.PeerManager), new(*domain.PeerManager)),
		wire.Bind(new(queries.PeerManager), new(*domain.PeerManager)),

		newBadger,

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

		invites.NewInviteRedeemer,
		wire.Bind(new(commands.InviteRedeemer), new(*invites.InviteRedeemer)),

		commands.NewTransactionRawMessagePublisher,
		wire.Bind(new(commands.RawMessagePublisher), new(*commands.TransactionRawMessagePublisher)),

		newContextLogger,

		portsSet,
		applicationSet,
		replicatorSet,
		blobReplicatorSet,
		formatsSet,
		pubSubSet,
		badgerNoTxRepositoriesSet,
		badgerTransactionProviderSet,
		badgerNoTxTransactionProviderSet,
		badgerAdaptersSet,
		blobsAdaptersSet,
		adaptersSet,
		extractFromConfigSet,
		networkingSet,
		migrationsSet,
		contentSet,
	)
	return IntegrationTestsService{}, nil, nil
}

var replicatorSet = wire.NewSet(
	gossip.NewManager,
	wire.Bind(new(gossip.ReplicationManager), new(*gossip.Manager)),

	gossip.NewGossipReplicator,
	wire.Bind(new(replication.CreateHistoryStreamReplicator), new(*gossip.GossipReplicator)),
	wire.Bind(new(ebt.SelfCreateHistoryStreamReplicator), new(*gossip.GossipReplicator)),

	ebt.NewReplicator,
	wire.Bind(new(replication.EpidemicBroadcastTreesReplicator), new(ebt.Replicator)),

	replication.NewWantedFeedsCache,
	wire.Bind(new(gossip.ContactsStorage), new(*replication.WantedFeedsCache)),
	wire.Bind(new(ebt.ContactsStorage), new(*replication.WantedFeedsCache)),

	ebt.NewSessionTracker,
	wire.Bind(new(ebt.Tracker), new(*ebt.SessionTracker)),

	ebt.NewSessionRunner,
	wire.Bind(new(ebt.Runner), new(*ebt.SessionRunner)),

	replication.NewNegotiator,
	wire.Bind(new(domain.MessageReplicator), new(*replication.Negotiator)),
)

func newAdvertiser(l identity.Public, config Config) (*local.Advertiser, error) {
	return local.NewAdvertiser(l, config.ListenAddress)
}

func newIntegrationTestConfig(t *testing.T) Config {
	dataDirectory := fixtures.Directory(t)
	oldDataDirectory := fixtures.Directory(t)

	cfg := Config{
		DataDirectory:      dataDirectory,
		GoSSBDataDirectory: oldDataDirectory,
		NetworkKey:         fixtures.SomeNetworkKey(),
		MessageHMAC:        fixtures.SomeMessageHMAC(),
	}
	cfg.SetDefaults()
	return cfg
}

func newBadger(system logging.LoggingSystem, logger logging.Logger, config Config) (*badger.DB, func(), error) {
	badgerDirectory := filepath.Join(config.DataDirectory, "badger")

	options := badger.DefaultOptions(badgerDirectory)
	options.Logger = badgeradapters.NewLogger(system, badgeradapters.LoggerLevelWarning)

	if config.ModifyBadgerOptions != nil {
		adapter := NewBadgerOptionsAdapter(&options)
		config.ModifyBadgerOptions(adapter)
	}

	db, err := badger.Open(options)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to open the database")
	}

	return db, func() {
		if err := db.Close(); err != nil {
			logger.WithError(err).Error("error closing the database")
		}
	}, nil
}

func privateIdentityToPublicIdentity(p identity.Private) identity.Public {
	return p.Public()
}

func newContextLogger(loggingSystem logging.LoggingSystem) logging.Logger {
	return logging.NewContextLogger(loggingSystem, "scuttlego")
}
