// Package rpc implements handlers for incoming Secure Scuttlebutt RPC requests.
package rpc

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc/mux"
)

// NewMuxHandlers is a convenience function used to create a list of all
// handlers implemented by this program.
func NewMuxHandlers(
	createHistoryStream *HandlerCreateHistoryStream,
	blobsGet *HandlerBlobsGet,
) []mux.Handler {
	return []mux.Handler{
		createHistoryStream,
		blobsGet,
	}
}

// NewMux is a convenience function which creates a new mux initialized with the
// provided list of handlers.
func NewMux(logger logging.Logger, handlers []mux.Handler) (*mux.Mux, error) {
	m := mux.NewMux(logger)

	for _, handler := range handlers {
		if err := m.AddHandler(handler); err != nil {
			return nil, errors.Wrap(err, "could not add a handler to the mux")
		}
	}

	return m, nil
}
