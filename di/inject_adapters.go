package di

import (
	"path"

	"github.com/google/wire"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/adapters"
	"github.com/planetary-social/go-ssb/service/adapters/blobs"
	"github.com/planetary-social/go-ssb/service/adapters/bolt"
	"github.com/planetary-social/go-ssb/service/adapters/mocks"
	"github.com/planetary-social/go-ssb/service/app/commands"
	"github.com/planetary-social/go-ssb/service/app/queries"
	blobReplication "github.com/planetary-social/go-ssb/service/domain/blobs/replication"
	"github.com/planetary-social/go-ssb/service/domain/feeds/formats"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/replication"
	"go.etcd.io/bbolt"
)

//nolint:deadcode,varcheck
var txBoltAdaptersSet = wire.NewSet(
	bolt.NewFeedRepository,
	wire.Bind(new(commands.FeedRepository), new(*bolt.FeedRepository)),

	bolt.NewSocialGraphRepository,
	wire.Bind(new(commands.SocialGraphRepository), new(*bolt.SocialGraphRepository)),

	bolt.NewWantListRepository,
	wire.Bind(new(commands.WantListRepository), new(*bolt.WantListRepository)),

	bolt.NewReceiveLogRepository,
	bolt.NewMessageRepository,
	bolt.NewPubRepository,
	bolt.NewBlobRepository,
)

//nolint:deadcode,varcheck
var boltAdaptersSet = wire.NewSet(
	bolt.NewReadFeedRepository,
	wire.Bind(new(queries.FeedRepository), new(*bolt.ReadFeedRepository)),

	bolt.NewReadContactsRepository,
	wire.Bind(new(replication.Storage), new(*bolt.ReadContactsRepository)),

	bolt.NewReadReceiveLogRepository,
	wire.Bind(new(queries.ReceiveLogRepository), new(*bolt.ReadReceiveLogRepository)),

	bolt.NewReadMessageRepository,
	wire.Bind(new(queries.MessageRepository), new(*bolt.ReadMessageRepository)),

	bolt.NewReadWantListRepository,
	wire.Bind(new(blobReplication.WantListStorage), new(*bolt.ReadWantListRepository)),

	newTxRepositoriesFactory,
)

//nolint:deadcode,varcheck
var mockQueryAdaptersSet = wire.NewSet(
	mocks.NewFeedRepositoryMock,
	wire.Bind(new(queries.FeedRepository), new(*mocks.FeedRepositoryMock)),

	mocks.NewReceiveLogRepositoryMock,
	wire.Bind(new(queries.ReceiveLogRepository), new(*mocks.ReceiveLogRepositoryMock)),

	mocks.NewMessageRepositoryMock,
	wire.Bind(new(queries.MessageRepository), new(*mocks.MessageRepositoryMock)),
)

func newTxRepositoriesFactory(local identity.Public, logger logging.Logger, hmac formats.MessageHMAC) bolt.TxRepositoriesFactory {
	return func(tx *bbolt.Tx) (bolt.TxRepositories, error) {
		return BuildTxRepositories(tx, local, logger, hmac)
	}
}

//nolint:deadcode,varcheck
var blobsAdaptersSet = wire.NewSet(
	newFilesystemStorage,
	wire.Bind(new(blobReplication.BlobStorage), new(*blobs.FilesystemStorage)),
	wire.Bind(new(queries.BlobStorage), new(*blobs.FilesystemStorage)),
	wire.Bind(new(blobReplication.BlobSizeRepository), new(*blobs.FilesystemStorage)),
)

func newFilesystemStorage(logger logging.Logger, config Config) (*blobs.FilesystemStorage, error) {
	return blobs.NewFilesystemStorage(path.Join(config.DataDirectory, "blobs"), logger)
}

//nolint:deadcode,varcheck
var adaptersSet = wire.NewSet(
	adapters.NewCurrentTimeProvider,
	wire.Bind(new(commands.CurrentTimeProvider), new(*adapters.CurrentTimeProvider)),
)
