// Package rpc implements handlers for incoming Secure Scuttlebutt RPC requests.
package rpc

import (
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux"
)

// NewMuxHandlers is a convenience function used to create a list of all
// handlers implemented by this program.
func NewMuxHandlers(
	blobsGet *HandlerBlobsGet,
	blobsCreateWants *HandlerBlobsCreateWants,
) []mux.Handler {
	return []mux.Handler{
		blobsGet,
		blobsCreateWants,
	}
}

// NewMuxClosingHandlers is a convenience function used to create a list of all
// handlers implemented by this program.
func NewMuxClosingHandlers(
	createHistoryStream *HandlerCreateHistoryStream,
) []mux.SynchronousHandler {
	return []mux.SynchronousHandler{
		createHistoryStream,
	}
}
