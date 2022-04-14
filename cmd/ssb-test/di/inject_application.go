package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/go-ssb/service/app"
	"github.com/planetary-social/go-ssb/service/app/commands"
	"github.com/planetary-social/go-ssb/service/app/queries"
	portsrpc "github.com/planetary-social/go-ssb/service/ports/rpc"
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
)

var queriesSet = wire.NewSet(
	wire.Struct(new(app.Queries), "*"),

	queries.NewCreateHistoryStreamHandler,
	wire.Bind(new(portsrpc.CreateHistoryStreamQueryHandler), new(*queries.CreateHistoryStreamHandler)),

	queries.NewGetReceiveLogHandler,
	queries.NewStatsHandler,
)