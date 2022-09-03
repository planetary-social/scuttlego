package di

import (
	"context"

	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/app"
	"github.com/planetary-social/scuttlego/service/domain/network/local"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux"
	portsnetwork "github.com/planetary-social/scuttlego/service/ports/network"
	portspubsub "github.com/planetary-social/scuttlego/service/ports/pubsub"
	portsrpc "github.com/planetary-social/scuttlego/service/ports/rpc"
)

//nolint:deadcode,varcheck
var portsSet = wire.NewSet(
	mux.NewMux,

	portsrpc.NewMuxHandlers,
	portsrpc.NewHandlerBlobsGet,
	portsrpc.NewHandlerBlobsCreateWants,

	portsrpc.NewMuxClosingHandlers,
	portsrpc.NewHandlerCreateHistoryStream,

	portspubsub.NewRequestSubscriber,

	local.NewDiscoverer,
	portsnetwork.NewDiscoverer,
	portsnetwork.NewConnectionEstablisher,

	newListener,
)

func newListener(
	ctx context.Context,
	initializer portsnetwork.ServerPeerInitializer,
	app app.Application,
	config Config,
	logger logging.Logger,
) (*portsnetwork.Listener, error) {
	return portsnetwork.NewListener(ctx, initializer, app, config.ListenAddress, logger)
}
