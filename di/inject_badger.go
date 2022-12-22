package di

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/google/wire"
	badgeradapters "github.com/planetary-social/scuttlego/service/adapters/badger"
)

//nolint:unused
var testBadgerTransactionProviderSet = wire.NewSet(
	badgeradapters.NewTestTransactionProvider,
	testAdaptersFactory,
)

func testAdaptersFactory() badgeradapters.TestAdaptersFactory {
	return func(tx *badger.Txn) (badgeradapters.TestAdapters, error) {
		return buildBadgerTestAdapters(tx)
	}
}
