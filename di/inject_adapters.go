package di

import (
	"path"

	"github.com/google/wire"
	mocks2 "github.com/planetary-social/scuttlego/internal/mocks"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/adapters"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/adapters/blobs"
	ebtadapters "github.com/planetary-social/scuttlego/service/adapters/ebt"
	invitesadapters "github.com/planetary-social/scuttlego/service/adapters/invites"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	blobreplication "github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/invites"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
)

var mockQueryAdaptersSet = wire.NewSet(
	mocks2.NewFeedRepositoryMock,
	wire.Bind(new(queries.FeedRepository), new(*mocks2.FeedRepositoryMock)),

	mocks2.NewReceiveLogRepositoryMock,
	wire.Bind(new(queries.ReceiveLogRepository), new(*mocks2.ReceiveLogRepositoryMock)),

	mocks2.NewMessageRepositoryMock,
	wire.Bind(new(queries.MessageRepository), new(*mocks2.MessageRepositoryMock)),

	mocks2.NewSocialGraphRepositoryMock,
	wire.Bind(new(queries.SocialGraphRepository), new(*mocks2.SocialGraphRepositoryMock)),

	mocks2.NewFeedWantListRepositoryMock,
	wire.Bind(new(queries.FeedWantListRepository), new(*mocks2.FeedWantListRepositoryMock)),

	mocks2.NewBanListRepositoryMock,
	wire.Bind(new(queries.BanListRepository), new(*mocks2.BanListRepositoryMock)),
)

var blobsAdaptersSet = wire.NewSet(
	newFilesystemStorage,
	wire.Bind(new(blobreplication.BlobStorage), new(*blobs.FilesystemStorage)),
	wire.Bind(new(blobreplication.BlobStorer), new(*blobs.FilesystemStorage)),
	wire.Bind(new(queries.BlobStorage), new(*blobs.FilesystemStorage)),
	wire.Bind(new(blobreplication.BlobSizeRepository), new(*blobs.FilesystemStorage)),
	wire.Bind(new(commands.BlobCreator), new(*blobs.FilesystemStorage)),
)

func newFilesystemStorage(logger logging.Logger, config Config) (*blobs.FilesystemStorage, error) {
	return blobs.NewFilesystemStorage(path.Join(config.GoSSBDataDirectory, "blobs"), logger)
}

var adaptersSet = wire.NewSet(
	adapters.NewCurrentTimeProvider,
	wire.Bind(new(commands.CurrentTimeProvider), new(*adapters.CurrentTimeProvider)),
	wire.Bind(new(boxstream.CurrentTimeProvider), new(*adapters.CurrentTimeProvider)),
	wire.Bind(new(invitesadapters.CurrentTimeProvider), new(*adapters.CurrentTimeProvider)),
	wire.Bind(new(blobreplication.CurrentTimeProvider), new(*adapters.CurrentTimeProvider)),

	adapters.NewBanListHasher,
	wire.Bind(new(badger.BanListHasher), new(*adapters.BanListHasher)),

	ebtadapters.NewCreateHistoryStreamHandlerAdapter,
	wire.Bind(new(ebt.MessageStreamer), new(*ebtadapters.CreateHistoryStreamHandlerAdapter)),

	invitesadapters.NewInviteDialer,
	wire.Bind(new(invites.InviteDialer), new(*invitesadapters.InviteDialer)),
)
