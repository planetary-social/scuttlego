package replication

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type BlobStorer interface {
	Store(id refs.Blob, r io.Reader) error
}

type BlobsGetDownloader struct {
	storer BlobStorer
	logger logging.Logger
}

func NewBlobsGetDownloader(
	storer BlobStorer,
	logger logging.Logger,
) *BlobsGetDownloader {
	return &BlobsGetDownloader{
		storer: storer,
		logger: logger.New("downloader"),
	}
}

func (d *BlobsGetDownloader) Download(ctx context.Context, peer transport.Peer, blob refs.Blob) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	arguments, err := messages.NewBlobsGetArguments(blob, nil, nil)
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

	go d.storeBlob(blob, pipeReader)

	if err := d.copyBlobContent(pipeWriter, rs); err != nil {
		return errors.Wrap(err, "failed to read blob content")
	}

	return nil
}

func (d *BlobsGetDownloader) copyBlobContent(pipeWriter *io.PipeWriter, rs rpc.ResponseStream) error {
	for chunk := range rs.Channel() {
		if err := chunk.Err; err != nil {
			if errors.Is(err, rpc.RemoteError{}) || errors.Is(err, rpc.ErrRemoteEnd) {
				break // todo can be an error or a real end of stream; is our rpc interface wrong?
				// todo interface has changed update?
			}
			err = errors.Wrap(err, "received an error")
			pipeWriter.CloseWithError(err) // nolint:errcheck, always returns nil
			return err
		}

		if _, err := pipeWriter.Write(chunk.Value.Bytes()); err != nil {
			err = errors.Wrap(err, "could not write to the pipe")
			pipeWriter.CloseWithError(err) // nolint:errcheck, always returns nil
			return err
		}
	}

	if err := pipeWriter.Close(); err != nil {
		return errors.Wrap(err, "could not close the pipe")
	}

	return nil
}

func (d *BlobsGetDownloader) storeBlob(id refs.Blob, r io.ReadCloser) {
	err := d.storer.Store(id, r)
	if err != nil {
		d.logger.WithError(err).Error("failed to save a blob")
	}
}
