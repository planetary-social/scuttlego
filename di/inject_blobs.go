package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/app/commands"
	blobReplication "github.com/planetary-social/scuttlego/service/domain/blobs/replication"
)

var blobReplicatorSet = wire.NewSet(
	blobReplication.NewManager,
	wire.Bind(new(blobReplication.ReplicationManager), new(*blobReplication.Manager)),
	wire.Bind(new(commands.BlobReplicationManager), new(*blobReplication.Manager)),

	newManagedWantsProcessFactory,
	wire.Bind(new(blobReplication.ManagedWantsProcessFactory), new(*managedWantsProcessFactory)),

	blobReplication.NewReplicator,
	wire.Bind(new(commands.BlobReplicator), new(*blobReplication.Replicator)),

	blobReplication.NewBlobsGetDownloader,
	wire.Bind(new(blobReplication.Downloader), new(*blobReplication.BlobsGetDownloader)),

	blobReplication.NewHasHandler,
	wire.Bind(new(blobReplication.HasBlobHandler), new(*blobReplication.HasHandler)),

	blobReplication.NewStorageBlobsThatShouldBePushedProvider,
	wire.Bind(new(blobReplication.BlobsThatShouldBePushedProvider), new(*blobReplication.StorageBlobsThatShouldBePushedProvider)),
)

type managedWantsProcessFactory struct {
	wantedBlobsProvider             blobReplication.WantedBlobsProvider
	blobsThatShouldBePushedProvider blobReplication.BlobsThatShouldBePushedProvider
	blobStorage                     blobReplication.BlobSizeRepository
	hasHandler                      blobReplication.HasBlobHandler
	logger                          logging.Logger
}

func newManagedWantsProcessFactory(
	wantedBlobsProvider blobReplication.WantedBlobsProvider,
	blobsThatShouldBePushedProvider blobReplication.BlobsThatShouldBePushedProvider,
	blobStorage blobReplication.BlobSizeRepository,
	hasHandler blobReplication.HasBlobHandler,
	logger logging.Logger,
) *managedWantsProcessFactory {
	return &managedWantsProcessFactory{
		wantedBlobsProvider:             wantedBlobsProvider,
		blobsThatShouldBePushedProvider: blobsThatShouldBePushedProvider,
		blobStorage:                     blobStorage,
		hasHandler:                      hasHandler,
		logger:                          logger,
	}
}

func (m managedWantsProcessFactory) NewWantsProcess() blobReplication.ManagedWantsProcess {
	return blobReplication.NewWantsProcess(
		m.wantedBlobsProvider,
		m.blobsThatShouldBePushedProvider,
		m.blobStorage,
		m.hasHandler,
		m.logger,
	)
}
