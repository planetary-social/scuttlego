package di

import (
	"path"

	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/adapters"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/adapters/blobs"
	ebtadapters "github.com/planetary-social/scuttlego/service/adapters/ebt"
	invitesadapters "github.com/planetary-social/scuttlego/service/adapters/invites"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	blobReplication "github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/invites"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
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
	return blobs.NewFilesystemStorage(path.Join(config.GoSSBDataDirectory, "blobs"), logger)
}

//nolint:unused
var adaptersSet = wire.NewSet(
	adapters.NewCurrentTimeProvider,
	wire.Bind(new(commands.CurrentTimeProvider), new(*adapters.CurrentTimeProvider)),
	wire.Bind(new(boxstream.CurrentTimeProvider), new(*adapters.CurrentTimeProvider)),
	wire.Bind(new(invitesadapters.CurrentTimeProvider), new(*adapters.CurrentTimeProvider)),

	adapters.NewBanListHasher,
	wire.Bind(new(badger.BanListHasher), new(*adapters.BanListHasher)),

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
)
