package rpc

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc/mux"
)

func NewMuxHandlers(
	createHistoryStream *HandlerCreateHistoryStream,
	blobsGet *HandlerBlobsGet,
) []mux.Handler {
	return []mux.Handler{
		createHistoryStream,
		blobsGet,
	}
}

func NewMux(logger logging.Logger, handlers []mux.Handler) (*mux.Mux, error) {
	mux := mux.NewMux(logger)

	for _, handler := range handlers {
		if err := mux.AddHandler(handler); err != nil {
			return nil, errors.Wrap(err, "could not add a handler to the mux")
		}
	}

	return mux, nil
}
