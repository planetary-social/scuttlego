package replication

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type BlobStorage interface {
	Has(id refs.Blob) (bool, error)
}

type WantListRepository interface {
	Contains(id refs.Blob) (bool, error)
	Delete(id refs.Blob) error
}

type Downloader interface {
	Download(ctx context.Context, peer transport.Peer, blob refs.Blob) error
}

type BlobDownloadedPublisher interface {
	Publish(blob refs.Blob, size blobs.Size)
}

type HasHandler struct {
	storage    BlobStorage
	wantList   WantListRepository
	downloader Downloader
	publisher  BlobDownloadedPublisher
	logger     logging.Logger
}

func NewHasHandler(
	storage BlobStorage,
	wantList WantListRepository,
	downloader Downloader,
	publisher BlobDownloadedPublisher,
	logger logging.Logger,
) *HasHandler {
	return &HasHandler{
		storage:    storage,
		wantList:   wantList,
		downloader: downloader,
		publisher:  publisher,
		logger:     logger.New("has_handler"),
	}
}

func (d *HasHandler) OnHasReceived(ctx context.Context, peer transport.Peer, blob refs.Blob, size blobs.Size) {
	d.logger.Debug().WithField("blob", blob.String()).Message("has")

	if err := d.onHasReceived(ctx, peer, blob, size); err != nil {
		d.logger.Error().WithError(err).Message("failed to download a blob")
	}
}

func (d *HasHandler) onHasReceived(ctx context.Context, peer transport.Peer, blob refs.Blob, declaredSize blobs.Size) error {
	if declaredSize.Above(blobs.MaxBlobSize()) {
		return errors.New("blob too large")
	}

	blobIsInTheWantList, err := d.wantList.Contains(blob)
	if err != nil {
		return errors.Wrap(err, "failed to check the want list")
	}

	if !blobIsInTheWantList {
		return nil
	}

	blobIsInStorage, err := d.storage.Has(blob)
	if err != nil {
		return errors.Wrap(err, "failed to check the storage")
	}

	if !blobIsInStorage {
		if err := d.downloader.Download(ctx, peer, blob); err != nil {
			return errors.Wrap(err, "download failed")
		}

		d.publisher.Publish(blob, declaredSize) // todo use actual size
	}

	if err := d.wantList.Delete(blob); err != nil {
		return errors.Wrap(err, "error deleting from want list")
	}

	return nil
}
