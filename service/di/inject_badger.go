package di

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/google/wire"
	mocks2 "github.com/planetary-social/scuttlego/internal/mocks"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service"
	badgeradapters "github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/adapters/badger/notx"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	blobReplication "github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/identity"
)

var badgerUnpackTestDependenciesSet = wire.NewSet(
	wire.FieldsOf(new(badgeradapters.TestAdaptersDependencies),
		"BanListHasher",
		"CurrentTimeProvider",
		"RawMessageIdentifier",
		"LocalIdentity",
	),
	wire.Bind(new(badgeradapters.BanListHasher), new(*mocks2.BanListHasherMock)),
	wire.Bind(new(commands.CurrentTimeProvider), new(*mocks2.CurrentTimeProviderMock)),
	wire.Bind(new(badgeradapters.RawMessageIdentifier), new(*mocks2.RawMessageIdentifierMock)),
)

var badgerAdaptersSet = wire.NewSet(
	badgeradapters.NewGarbageCollector,
)

var badgerNoTxRepositoriesSet = wire.NewSet(
	notx.NewNoTxBlobWantListRepository,
	wire.Bind(new(blobReplication.WantedBlobsProvider), new(*notx.NoTxBlobWantListRepository)),
	wire.Bind(new(blobReplication.WantListRepository), new(*notx.NoTxBlobWantListRepository)),

	notx.NewNoTxBlobsRepository,
	wire.Bind(new(blobReplication.BlobsRepository), new(*notx.NoTxBlobsRepository)),

	notx.NewNoTxFeedWantListRepository,
)

var badgerRepositoriesSet = wire.NewSet(
	badgeradapters.NewBanListRepository,
	wire.Bind(new(commands.BanListRepository), new(*badgeradapters.BanListRepository)),
	wire.Bind(new(queries.BanListRepository), new(*badgeradapters.BanListRepository)),

	badgeradapters.NewBlobWantListRepository,
	wire.Bind(new(commands.BlobWantListRepository), new(*badgeradapters.BlobWantListRepository)),
	wire.Bind(new(blobReplication.WantListRepository), new(*badgeradapters.BlobWantListRepository)),

	badgeradapters.NewFeedWantListRepository,
	wire.Bind(new(commands.FeedWantListRepository), new(*badgeradapters.FeedWantListRepository)),
	wire.Bind(new(queries.FeedWantListRepository), new(*badgeradapters.FeedWantListRepository)),

	badgeradapters.NewReceiveLogRepository,
	wire.Bind(new(commands.ReceiveLogRepository), new(*badgeradapters.ReceiveLogRepository)),
	wire.Bind(new(queries.ReceiveLogRepository), new(*badgeradapters.ReceiveLogRepository)),

	badgeradapters.NewSocialGraphRepository,
	wire.Bind(new(commands.SocialGraphRepository), new(*badgeradapters.SocialGraphRepository)),
	wire.Bind(new(queries.SocialGraphRepository), new(*badgeradapters.SocialGraphRepository)),

	badgeradapters.NewFeedRepository,
	wire.Bind(new(commands.FeedRepository), new(*badgeradapters.FeedRepository)),
	wire.Bind(new(queries.FeedRepository), new(*badgeradapters.FeedRepository)),

	badgeradapters.NewMessageRepository,
	wire.Bind(new(queries.MessageRepository), new(*badgeradapters.MessageRepository)),

	badgeradapters.NewPubRepository,
	badgeradapters.NewBlobRepository,
)

var badgerTestAdaptersDependenciesSet = wire.NewSet(
	wire.Struct(new(badgeradapters.TestAdaptersDependencies), "*"),
	mocks2.NewBanListHasherMock,
	mocks2.NewCurrentTimeProviderMock,
	mocks2.NewRawMessageIdentifierMock,
)

var badgerNoTxTestTransactionProviderSet = wire.NewSet(
	notx.NewTestTxAdaptersFactoryTransactionProvider,
	wire.Bind(new(notx.TransactionProvider), new(*notx.TestTxAdaptersFactoryTransactionProvider)),

	noTxTestTxAdaptersFactory,
)

var testBadgerTransactionProviderSet = wire.NewSet(
	badgeradapters.NewTestTransactionProvider,
	testAdaptersFactory,
)

var badgerTransactionProviderSet = wire.NewSet(
	badgeradapters.NewCommandsTransactionProvider,
	wire.Bind(new(commands.TransactionProvider), new(*badgeradapters.CommandsTransactionProvider)),

	badgerCommandsAdaptersFactory,

	badgeradapters.NewQueriesTransactionProvider,
	wire.Bind(new(queries.TransactionProvider), new(*badgeradapters.QueriesTransactionProvider)),

	badgerQueriesAdaptersFactory,
)

var badgerNoTxTransactionProviderSet = wire.NewSet(
	notx.NewTxAdaptersFactoryTransactionProvider,
	wire.Bind(new(notx.TransactionProvider), new(*notx.TxAdaptersFactoryTransactionProvider)),

	noTxTxAdaptersFactory,
)

func noTxTestTxAdaptersFactory() notx.TestTxAdaptersFactory {
	return func(tx *badger.Txn, dependencies badgeradapters.TestAdaptersDependencies) (notx.TxAdapters, error) {
		return buildTestBadgerNoTxTxAdapters(tx, dependencies)
	}
}

func noTxTxAdaptersFactory(local identity.Public, conf service.Config, logger logging.Logger) notx.TxAdaptersFactory {
	return func(tx *badger.Txn) (notx.TxAdapters, error) {
		return buildBadgerNoTxTxAdapters(tx, local, conf, logger)
	}
}

func testAdaptersFactory() badgeradapters.TestAdaptersFactory {
	return func(tx *badger.Txn, dependencies badgeradapters.TestAdaptersDependencies) (badgeradapters.TestAdapters, error) {
		return buildBadgerTestAdapters(tx, dependencies)
	}
}

func badgerCommandsAdaptersFactory(config service.Config, local identity.Public, logger logging.Logger) badgeradapters.CommandsAdaptersFactory {
	return func(tx *badger.Txn) (commands.Adapters, error) {
		return buildBadgerCommandsAdapters(tx, local, config, logger)
	}
}

func badgerQueriesAdaptersFactory(config service.Config, local identity.Public, logger logging.Logger) badgeradapters.QueriesAdaptersFactory {
	return func(tx *badger.Txn) (queries.Adapters, error) {
		return buildBadgerQueriesAdapters(tx, local, config, logger)
	}
}
