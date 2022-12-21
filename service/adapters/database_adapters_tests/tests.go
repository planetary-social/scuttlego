package badger

import (
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/bans"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"reflect"
	"runtime"
	"testing"
)

func Run(t *testing.T, testedSystems []*TestedSystem) {
	for _, testedSystem := range testedSystems {
		for _, testFunction := range testFunctions {
			testFunctionName := runtime.FuncForPC(reflect.ValueOf(testFunction).Pointer()).Name()
			t.Run(testFunctionName, func(t *testing.T) {
				testFunction(t, testedSystem)
			})
		}
	}
}

var testFunctions = []testFunction{
	testBanListRepository_LookupMappingReturnsCorrectErrorWhenMappingDoesNotExist,
	testBanListRepository_CreateFeedMappingInsertsMappingsAndRemoveFeedMappingRemovesMappings,
	testBanListRepository_AddInsertsHashesAndRemoveRemovesHashesFromBanList,
	testBanListRepository_ContainsFeedCorrectlyLooksUpHashes,
}

type testFunction func(t *testing.T, ts *TestedSystem)

type TestedSystem struct {
	TransactionProvider TransactionProvider
}

type TransactionProvider interface {
	Update(f func(adapters Adapters) error) error
	View(f func(adapters Adapters) error) error
}

type Adapters struct {
	BanList BanListRepository

	BanListHasher *mocks.BanListHasherMock
}

type BanListRepository interface {
	Add(hash bans.Hash) error
	Remove(hash bans.Hash) error
	Contains(hash bans.Hash) (bool, error)
	ContainsFeed(ref refs.Feed) (bool, error)
	CreateFeedMapping(ref refs.Feed) error
	RemoveFeedMapping(ref refs.Feed) error
	LookupMapping(hash bans.Hash) (commands.BannableRef, error)
}
