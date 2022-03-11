package rpc

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
)

func NewMuxHandlers(
	createHistoryStream *HandlerCreateHistoryStream,
	blobsGet *HandlerBlobsGet,
) []rpc.Handler {
	return []rpc.Handler{
		createHistoryStream,
		blobsGet,
	}
}

func NewMux(logger logging.Logger, handlers []rpc.Handler) (*rpc.Mux, error) {
	mux := rpc.NewMux(logger)

	for _, handler := range handlers {
		if err := mux.AddHandler(handler); err != nil {
			return nil, errors.Wrap(err, "could not add a handler to the mux")
		}
	}

	return mux, nil
}
