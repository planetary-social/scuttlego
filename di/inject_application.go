package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/service/app"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	portsrpc "github.com/planetary-social/scuttlego/service/ports/rpc"
)

//nolint:deadcode,varcheck
var applicationSet = wire.NewSet(
	wire.Struct(new(app.Application), "*"),

	commandsSet,
	queriesSet,
)

var commandsSet = wire.NewSet(
	wire.Struct(new(app.Commands), "*"),

	commands.NewRedeemInviteHandler,
	commands.NewFollowHandler,
	commands.NewConnectHandler,
	commands.NewAcceptNewPeerHandler,
	commands.NewProcessNewLocalDiscoveryHandler,
	commands.NewPublishRawHandler,
	commands.NewEstablishNewConnectionsHandler,
	commands.NewDownloadBlobHandler,

	commands.NewRawMessageHandler,
	wire.Bind(new(replication.RawMessageHandler), new(*commands.RawMessageHandler)),

	commands.NewCreateWantsHandler,
	wire.Bind(new(portsrpc.CreateWantsCommandHandler), new(*commands.CreateWantsHandler)),
)

var queriesSet = wire.NewSet(
	wire.Struct(new(app.Queries), "*"),

	queries.NewCreateHistoryStreamHandler,
	wire.Bind(new(portsrpc.CreateHistoryStreamQueryHandler), new(*queries.CreateHistoryStreamHandler)),

	queries.NewReceiveLogHandler,
	queries.NewStatusHandler,
	queries.NewPublishedMessagesHandler,
	queries.NewBlobDownloadedEventsHandler,

	queries.NewGetBlobHandler,
	wire.Bind(new(portsrpc.GetBlobQueryHandler), new(*queries.GetBlobHandler)),
)
