package rpc

import (
	"bytes"
	"context"
	"io"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/app/queries"
	"github.com/planetary-social/go-ssb/service/domain/messages"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc/mux"
)

const (
	// MaxBlobChunkSizeInBytes specifies the max size of one response payload
	// when sending blobs. Larger blobs are broken into several payloads. This
	// is done to give other pieces of code some wire time and avoid blocking
	// the connection completely when sending the blob.
	MaxBlobChunkSizeInBytes = 100 * 1000
)

type GetBlobQueryHandler interface {
	Handle(query queries.GetBlob) (io.ReadCloser, error)
}

type HandlerBlobsGet struct {
	handler GetBlobQueryHandler
}

func NewHandlerBlobsGet(handler GetBlobQueryHandler) *HandlerBlobsGet {
	return &HandlerBlobsGet{
		handler: handler,
	}
}

func (h HandlerBlobsGet) Procedure() rpc.Procedure {
	return messages.BlobsGetProcedure
}

func (h HandlerBlobsGet) Handle(ctx context.Context, w mux.ResponseWriter, req *rpc.Request) error {
	args, err := messages.NewBlobsGetArgumentsFromBytes(req.Arguments())
	if err != nil {
		return errors.Wrap(err, "invalid arguments")
	}

	query := queries.GetBlob{
		Id: args.Id(),
	}

	rc, err := h.handler.Handle(query)
	if err != nil {
		return errors.Wrap(err, "error executing the query")
	}
	defer rc.Close()

	buf := &bytes.Buffer{}

	for {
		n, err := io.Copy(buf, io.LimitReader(rc, MaxBlobChunkSizeInBytes))
		if err != nil {
			return errors.Wrap(err, "failed to copy into buffer")
		}

		if n == 0 {
			return nil
		}

		if err := w.WriteMessage(buf.Bytes()); err != nil {
			return errors.Wrap(err, "failed to write the message")
		}

		buf.Reset()
	}
}
