package di

import (
	"path"

	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/adapters"
	"github.com/planetary-social/scuttlego/service/adapters/blobs"
	"github.com/planetary-social/scuttlego/service/adapters/bolt"
	ebtadapters "github.com/planetary-social/scuttlego/service/adapters/ebt"
	invitesadapters "github.com/planetary-social/scuttlego/service/adapters/invites"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	blobReplication "github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/invites"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
	"go.etcd.io/bbolt"
)

//nolint:unused
var txBoltAdaptersSet = wire.NewSet(
	bolt.NewFeedRepository,
	wire.Bind(new(commands.FeedRepository), new(*bolt.FeedRepository)),

	bolt.NewSocialGraphRepository,
	wire.Bind(new(commands.SocialGraphRepository), new(*bolt.SocialGraphRepository)),

	bolt.NewBlobWantListRepository,
	wire.Bind(new(commands.BlobWantListRepository), new(*bolt.BlobWantListRepository)),
	wire.Bind(new(blobReplication.WantListRepository), new(*bolt.BlobWantListRepository)),

	bolt.NewFeedWantListRepository,
	wire.Bind(new(commands.FeedWantListRepository), new(*bolt.FeedWantListRepository)),

	bolt.NewBanListRepository,
	wire.Bind(new(commands.BanListRepository), new(*bolt.BanListRepository)),

	bolt.NewReceiveLogRepository,
	bolt.NewMessageRepository,
	bolt.NewPubRepository,
	bolt.NewBlobRepository,
	bolt.NewWantedFeedsRepository,
)

//nolint:unused
var boltAdaptersSet = wire.NewSet(
	bolt.NewReadFeedRepository,
	wire.Bind(new(queries.FeedRepository), new(*bolt.ReadFeedRepository)),

	bolt.NewReadWantedFeedsRepository,
	wire.Bind(new(replication.WantedFeedsRepository), new(*bolt.ReadWantedFeedsRepository)),

	bolt.NewReadReceiveLogRepository,
	wire.Bind(new(queries.ReceiveLogRepository), new(*bolt.ReadReceiveLogRepository)),

	bolt.NewReadMessageRepository,
	wire.Bind(new(queries.MessageRepository), new(*bolt.ReadMessageRepository)),

	bolt.NewReadBlobWantListRepository,
	wire.Bind(new(blobReplication.WantListStorage), new(*bolt.ReadBlobWantListRepository)),
	wire.Bind(new(blobReplication.WantListRepository), new(*bolt.ReadBlobWantListRepository)),

	newTxRepositoriesFactory,
)

//nolint:unused
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

//nolint:unused
var blobsAdaptersSet = wire.NewSet(
	newFilesystemStorage,
	wire.Bind(new(blobReplication.BlobStorage), new(*blobs.FilesystemStorage)),
	wire.Bind(new(blobReplication.BlobStorer), new(*blobs.FilesystemStorage)),
	wire.Bind(new(queries.BlobStorage), new(*blobs.FilesystemStorage)),
	wire.Bind(new(blobReplication.BlobSizeRepository), new(*blobs.FilesystemStorage)),
	wire.Bind(new(commands.BlobCreator), new(*blobs.FilesystemStorage)),
)

func newFilesystemStorage(logger logging.Logger, config Config) (*blobs.FilesystemStorage, error) {
	return blobs.NewFilesystemStorage(path.Join(config.DataDirectory, "blobs"), logger)
}

//nolint:unused
var adaptersSet = wire.NewSet(
	adapters.NewCurrentTimeProvider,
	wire.Bind(new(commands.CurrentTimeProvider), new(*adapters.CurrentTimeProvider)),
	wire.Bind(new(boxstream.CurrentTimeProvider), new(*adapters.CurrentTimeProvider)),
	wire.Bind(new(invitesadapters.CurrentTimeProvider), new(*adapters.CurrentTimeProvider)),

	adapters.NewBanListHasher,
	wire.Bind(new(bolt.BanListHasher), new(*adapters.BanListHasher)),

	ebtadapters.NewCreateHistoryStreamHandlerAdapter,
	wire.Bind(new(ebt.MessageStreamer), new(*ebtadapters.CreateHistoryStreamHandlerAdapter)),

	invitesadapters.NewInviteDialer,
	wire.Bind(new(invites.InviteDialer), new(*invitesadapters.InviteDialer)),
)

//nolint:unused
var testAdaptersSet = wire.NewSet(
	mocks.NewCurrentTimeProviderMock,
	wire.Bind(new(commands.CurrentTimeProvider), new(*mocks.CurrentTimeProviderMock)),

	mocks.NewBanListHasherMock,
	wire.Bind(new(bolt.BanListHasher), new(*mocks.BanListHasherMock)),
)
