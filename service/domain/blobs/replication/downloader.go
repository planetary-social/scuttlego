package replication

import (
	"context"
	"fmt"
	"io"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/blobs"
	"github.com/planetary-social/go-ssb/service/domain/messages"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/transport"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
)

type BlobStorage interface {
	Save(id refs.Blob, r io.Reader) error
}

type BlobsGetDownloader struct {
	storage BlobStorage
	logger  logging.Logger
}

func NewBlobsGetDownloader(storage BlobStorage, logger logging.Logger) *BlobsGetDownloader {
	return &BlobsGetDownloader{
		storage: storage,
		logger:  logger.New("downloader"),
	}
}

func (d *BlobsGetDownloader) OnHasReceived(ctx context.Context, peer transport.Peer, blob refs.Blob, size blobs.Size) {
	d.logger.WithField("blob", blob.String()).Debug("has")

	if size.Above(blobs.MaxBlobSize()) {
		d.logger.WithField("size", size).Debug("blob too large")
		return
	}

	// todo pre check if we want this blob

	if err := d.download(ctx, peer, blob); err != nil {
		d.logger.WithError(err).Error("failed to download a blob")
	}
}

func (d *BlobsGetDownloader) download(ctx context.Context, peer transport.Peer, blob refs.Blob) error {
	arguments, err := messages.NewBlobsGetArguments(blob)
	if err != nil {
		return errors.Wrap(err, "could not create a request")
	}

	request, err := messages.NewBlobsGet(arguments)
	if err != nil {
		return errors.Wrap(err, "could not create a request")
	}

	rs, err := peer.Conn().PerformRequest(ctx, request)
	if err != nil {
		fmt.Println("performing request error", err)
		return errors.Wrap(err, "could not perform a request")
	}

	pipeReader, pipeWriter := io.Pipe()

	go d.save(blob, pipeReader)

	for chunk := range rs.Channel() {
		if err := chunk.Err; err != nil {
			if errors.Is(err, rpc.ErrEndOrErr) {
				break // todo can be an error or a real end of stream; is our rpc interface wrong?
			}
			return errors.Wrap(err, "received an error")
		}

		if _, err := pipeWriter.Write(chunk.Value.Bytes()); err != nil {
			return errors.Wrap(err, "could not write to the pipe")
		}
	}

	if err := pipeWriter.Close(); err != nil {
		return errors.Wrap(err, "could not close the pipe")
	}

	return nil
}

func (d *BlobsGetDownloader) save(id refs.Blob, r io.ReadCloser) {
	err := d.storage.Save(id, r)
	if err != nil {
		d.logger.WithError(err).Error("failed to save a blob")
	}
}
