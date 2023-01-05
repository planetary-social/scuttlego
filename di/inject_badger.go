package di

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/google/wire"
	badgeradapters "github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/adapters/badger/notx"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/app/commands"
)

//nolint:unused
var badgerUnpackTestDependenciesSet = wire.NewSet(
	wire.FieldsOf(new(badgeradapters.TestAdaptersDependencies),
		"BanListHasher",
		"CurrentTimeProvider",
		"RawMessageIdentifier",
		"LocalIdentity",
	),
	wire.Bind(new(badgeradapters.BanListHasher), new(*mocks.BanListHasherMock)),
	wire.Bind(new(commands.CurrentTimeProvider), new(*mocks.CurrentTimeProviderMock)),
	wire.Bind(new(badgeradapters.RawMessageIdentifier), new(*mocks.RawMessageIdentifierMock)),
)

//nolint:unused
var badgerNoTxRepositoriesSet = wire.NewSet(
	notx.NewNoTxBlobWantListRepository,
	notx.NewNoTxFeedRepository,
	notx.NewNoTxMessageRepository,
	notx.NewNoTxReceiveLogRepository,
	notx.NewNoTxWantedFeedsRepository,
)

//nolint:unused
var badgerRepositoriesSet = wire.NewSet(
	badgeradapters.NewBanListRepository,
	badgeradapters.NewBlobRepository,
	badgeradapters.NewBlobWantListRepository,
	badgeradapters.NewFeedWantListRepository,
	badgeradapters.NewMessageRepository,
	badgeradapters.NewReceiveLogRepository,
	badgeradapters.NewSocialGraphRepository,
	badgeradapters.NewPubRepository,
	badgeradapters.NewFeedRepository,
	badgeradapters.NewWantedFeedsRepository,
)

//nolint:unused
var badgerTestAdaptersDependenciesSet = wire.NewSet(
	wire.Struct(new(badgeradapters.TestAdaptersDependencies), "*"),
	mocks.NewBanListHasherMock,
	mocks.NewCurrentTimeProviderMock,
	mocks.NewRawMessageIdentifierMock,
)

//nolint:unused
var badgerNoTxTestTransactionProviderSet = wire.NewSet(
	notx.NewTestTxAdaptersFactoryTransactionProvider,
	wire.Bind(new(notx.TransactionProvider), new(*notx.TestTxAdaptersFactoryTransactionProvider)),

	noTxTestTxAdaptersFactory,
)

//nolint:unused
var badgerNoTxTransactionProviderSet = wire.NewSet(
	notx.NewTxAdaptersFactoryTransactionProvider,
	noTxTxAdaptersFactory,
)

//nolint:unused
var testBadgerTransactionProviderSet = wire.NewSet(
	badgeradapters.NewTestTransactionProvider,
	testAdaptersFactory,
)

func noTxTestTxAdaptersFactory() notx.TestTxAdaptersFactory {
	return func(tx *badger.Txn, dependencies badgeradapters.TestAdaptersDependencies) (notx.TxAdapters, error) {
		return buildTestBadgerNoTxTxAdapters(tx, dependencies)
	}
}

func noTxTxAdaptersFactory() notx.TxAdaptersFactory {
	return func(tx *badger.Txn) (notx.TxAdapters, error) {
		return buildBadgerNoTxTxAdapters(tx)
	}
}

func testAdaptersFactory() badgeradapters.TestAdaptersFactory {
	return func(tx *badger.Txn, dependencies badgeradapters.TestAdaptersDependencies) (badgeradapters.TestAdapters, error) {
		return buildBadgerTestAdapters(tx, dependencies)
	}
}
