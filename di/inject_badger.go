package di

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/google/wire"
	badgeradapters "github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/adapters/badger/notx"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
)

//nolint:unused
var badgerNoTxRepositoriesSet = wire.NewSet(
	notx.NewNoTxBlobWantListRepository,
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
)

//nolint:unused
var badgerTestAdaptersDependenciesSet = wire.NewSet(
	wire.Struct(new(badgeradapters.TestAdaptersDependencies), "*"),
	mocks.NewBanListHasherMock,
	mocks.NewCurrentTimeProviderMock,
	mocks.NewRawMessageIdentifierMock,
)

//nolint:unused
var badgerNoTxTransactionProviderSet = wire.NewSet(
	notx.NewTransactionProvider,
	noTxTxAdaptersFactory,
)

//nolint:unused
var testBadgerTransactionProviderSet = wire.NewSet(
	badgeradapters.NewTestTransactionProvider,
	testAdaptersFactory,
)

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
