//go:build wireinject
// +build wireinject

package di

import (
	"path/filepath"
	"testing"

	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	mocks2 "github.com/planetary-social/scuttlego/internal/mocks"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service"
	badgeradapters "github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/adapters/badger/notx"
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
	"github.com/planetary-social/scuttlego/service/domain/network/local"
	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/rooms/tunnel"
)

type BadgerNoTxTestAdapters struct {
	NoTxTestAdapters    notx.TestAdapters
	TransactionProvider *badgeradapters.TestTransactionProvider
	Dependencies        *badgeradapters.TestAdaptersDependencies
}

func BuildBadgerNoTxTestAdapters(testing.TB) BadgerNoTxTestAdapters {
	wire.Build(
		wire.Struct(new(BadgerNoTxTestAdapters), "*"),

		wire.Struct(new(notx.TestAdapters), "*"),

		badgerNoTxTestTransactionProviderSet,
		badgerNoTxRepositoriesSet,
		testBadgerTransactionProviderSet,
		badgerTestAdaptersDependenciesSet,

		fixtures.SomePublicIdentity,
		fixtures.Badger,

		logging.NewDevNullLogger,
		wire.Bind(new(logging.Logger), new(logging.DevNullLogger)),
	)

	return BadgerNoTxTestAdapters{}
}

type BadgerTestAdapters struct {
	TransactionProvider *badgeradapters.TestTransactionProvider
	Dependencies        *badgeradapters.TestAdaptersDependencies
}

func BuildBadgerTestAdapters(testing.TB) BadgerTestAdapters {
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

func buildBadgerNoTxTxAdapters(*badger.Txn, identity.Public, service.Config, logging.Logger) (notx.TxAdapters, error) {
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

	PeerManager            *mocks2.PeerManagerMock
	Dialer                 *mocks2.DialerMock
	FeedWantListRepository *mocks2.FeedWantListRepositoryMock
	CurrentTimeProvider    *mocks2.CurrentTimeProviderMock
	InviteRedeemer         *mocks2.InviteRedeemerMock
	Local                  identity.Public
	PeerInitializer        *mocks2.PeerInitializerMock
	GoSSBRepoReader        *mocks2.GoSSBRepoReaderMock
	FeedRepository         *mocks2.FeedRepositoryMock
	ReceiveLog             *mocks2.ReceiveLogRepositoryMock
}

func BuildTestCommands(testing.TB) (TestCommands, error) {
	wire.Build(
		commandsSet,
		migrationCommandsSet,

		mocks2.NewDialerMock,
		wire.Bind(new(commands.Dialer), new(*mocks2.DialerMock)),

		mocks2.NewPeerManagerMock,
		wire.Bind(new(commands.PeerManager), new(*mocks2.PeerManagerMock)),

		identity.NewPrivate,
		privateIdentityToPublicIdentity,

		mocks2.NewMockCommandsTransactionProvider,
		wire.Bind(new(commands.TransactionProvider), new(*mocks2.MockCommandsTransactionProvider)),

		wire.Struct(
			new(commands.Adapters),
			"FeedWantList",
			"Feed",
			"ReceiveLog",
		),

		mocks2.NewFeedWantListRepositoryMock,
		wire.Bind(new(commands.FeedWantListRepository), new(*mocks2.FeedWantListRepositoryMock)),

		mocks2.NewCurrentTimeProviderMock,
		wire.Bind(new(commands.CurrentTimeProvider), new(*mocks2.CurrentTimeProviderMock)),

		mocks2.NewInviteRedeemerMock,
		wire.Bind(new(commands.InviteRedeemer), new(*mocks2.InviteRedeemerMock)),

		mocks2.NewPeerInitializerMock,
		wire.Bind(new(commands.ServerPeerInitializer), new(*mocks2.PeerInitializerMock)),

		mocks2.NewGoSSBRepoReaderMock,
		wire.Bind(new(commands.GoSSBRepoReader), new(*mocks2.GoSSBRepoReaderMock)),

		mocks2.NewContentParser,
		wire.Bind(new(commands.ContentParser), new(*mocks2.ContentParser)),

		mocks2.NewFeedRepositoryMock,
		wire.Bind(new(commands.FeedRepository), new(*mocks2.FeedRepositoryMock)),

		mocks2.NewReceiveLogRepositoryMock,
		wire.Bind(new(commands.ReceiveLogRepository), new(*mocks2.ReceiveLogRepositoryMock)),

		fixtures.TestLogger,

		wire.Struct(new(TestCommands), "*"),
	)

	return TestCommands{}, nil
}

type TestQueries struct {
	Queries app.Queries

	WantedFeedsProvider *queries.WantedFeedsProvider

	FeedRepository         *mocks2.FeedRepositoryMock
	MessageRepository      *mocks2.MessageRepositoryMock
	ReceiveLogRepository   *mocks2.ReceiveLogRepositoryMock
	SocialGraphRepository  *mocks2.SocialGraphRepositoryMock
	FeedWantListRepository *mocks2.FeedWantListRepositoryMock
	BanListRepository      *mocks2.BanListRepositoryMock
	MessagePubSub          *mocks2.MessagePubSubMock
	PeerManager            *mocks2.PeerManagerMock
	BlobStorage            *mocks2.BlobStorageMock
	Dialer                 *mocks2.DialerMock

	LocalIdentity identity.Public
}

func BuildTestQueries(testing.TB) (TestQueries, error) {
	wire.Build(
		applicationSet,
		mockQueryAdaptersSet,
		replicationSet,

		mocks2.NewMockQueriesTransactionProvider,
		wire.Bind(new(queries.TransactionProvider), new(*mocks2.MockQueriesTransactionProvider)),

		wire.Struct(new(queries.Adapters), "*"),

		pubsub.NewMessagePubSub,
		mocks2.NewMessagePubSubMock,
		wire.Bind(new(queries.MessageSubscriber), new(*mocks2.MessagePubSubMock)),

		mocks2.NewPeerManagerMock,
		wire.Bind(new(queries.PeerManager), new(*mocks2.PeerManagerMock)),

		identity.NewPrivate,
		privateIdentityToPublicIdentity,

		mocks2.NewBlobStorageMock,
		wire.Bind(new(queries.BlobStorage), new(*mocks2.BlobStorageMock)),

		mocks2.NewBlobDownloadedPubSubMock,
		wire.Bind(new(queries.BlobDownloadedSubscriber), new(*mocks2.BlobDownloadedPubSubMock)),

		mocks2.NewDialerMock,
		wire.Bind(new(queries.Dialer), new(*mocks2.DialerMock)),

		wire.Struct(new(TestQueries), "*"),

		fixtures.TestLogger,
	)

	return TestQueries{}, nil
}

func buildBadgerCommandsAdapters(*badger.Txn, identity.Public, service.Config, logging.Logger) (commands.Adapters, error) {
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

func buildBadgerQueriesAdapters(*badger.Txn, identity.Public, service.Config, logging.Logger) (queries.Adapters, error) {
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
func BuildService(identity.Private, service.Config) (service.Service, func(), error) {
	wire.Build(
		service.NewService,

		domain.NewPeerManager,
		wire.Bind(new(commands.PeerManager), new(*domain.PeerManager)),
		wire.Bind(new(queries.PeerManager), new(*domain.PeerManager)),

		newBadger,

		newAdvertiser,
		privateIdentityToPublicIdentity,

		commands.NewMessageBuffer,

		rooms.NewScanner,
		wire.Bind(new(commands.RoomScanner), new(*rooms.Scanner)),

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
		replicationSet,
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
	return service.Service{}, nil, nil
}

type IntegrationTestsService struct {
	Service service.Service

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

		service.NewService,

		domain.NewPeerManager,
		wire.Bind(new(commands.PeerManager), new(*domain.PeerManager)),
		wire.Bind(new(queries.PeerManager), new(*domain.PeerManager)),

		newBadger,

		newAdvertiser,
		privateIdentityToPublicIdentity,

		commands.NewMessageBuffer,

		rooms.NewScanner,
		wire.Bind(new(commands.RoomScanner), new(*rooms.Scanner)),

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
		replicationSet,
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

func newAdvertiser(l identity.Public, config service.Config) (*local.Advertiser, error) {
	return local.NewAdvertiser(l, config.ListenAddress)
}

func newIntegrationTestConfig(t *testing.T) service.Config {
	dataDirectory := fixtures.Directory(t)
	oldDataDirectory := fixtures.Directory(t)

	cfg := service.Config{
		DataDirectory:      dataDirectory,
		GoSSBDataDirectory: oldDataDirectory,
		NetworkKey:         fixtures.SomeNetworkKey(),
		MessageHMAC:        fixtures.SomeMessageHMAC(),
	}
	cfg.SetDefaults()
	return cfg
}

func newBadger(system logging.LoggingSystem, logger logging.Logger, config service.Config) (*badger.DB, func(), error) {
	badgerDirectory := filepath.Join(config.DataDirectory, "badger")

	options := badger.DefaultOptions(badgerDirectory)
	options.Logger = badgeradapters.NewLogger(system, badgeradapters.LoggerLevelWarning)

	if config.ModifyBadgerOptions != nil {
		adapter := service.NewBadgerOptionsAdapter(&options)
		config.ModifyBadgerOptions(adapter)
	}

	db, err := badger.Open(options)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to open the database")
	}

	return db, func() {
		if err := db.Close(); err != nil {
			logger.Error().WithError(err).Message("error closing the database")
		}
	}, nil
}

func privateIdentityToPublicIdentity(p identity.Private) identity.Public {
	return p.Public()
}

func newContextLogger(loggingSystem logging.LoggingSystem) logging.Logger {
	return logging.NewContextLogger(loggingSystem, "scuttlego")
}
