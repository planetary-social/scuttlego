package di

import (
	"github.com/google/wire"
	ebtadapters "github.com/planetary-social/scuttlego/service/adapters/ebt"
	"github.com/planetary-social/scuttlego/service/app"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/ports/pubsub"
	portsrpc "github.com/planetary-social/scuttlego/service/ports/rpc"
)

//nolint:unused
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
	commands.NewDisconnectAllHandler,
	commands.NewAcceptNewPeerHandler,
	commands.NewProcessNewLocalDiscoveryHandler,
	commands.NewPublishRawHandler,
	commands.NewEstablishNewConnectionsHandler,
	commands.NewDownloadBlobHandler,
	commands.NewCreateBlobHandler,

	commands.NewRawMessageHandler,
	wire.Bind(new(replication.RawMessageHandler), new(*commands.RawMessageHandler)),

	commands.NewCreateWantsHandler,
	wire.Bind(new(portsrpc.CreateWantsCommandHandler), new(*commands.CreateWantsHandler)),

	commands.NewAddToBanListHandler,
	commands.NewRemoveFromBanListHandler,

	commands.NewHandleIncomingEbtReplicateHandler,
	wire.Bind(new(portsrpc.EbtReplicateCommandHandler), new(*commands.HandleIncomingEbtReplicateHandler)),

	commands.NewRoomsAliasRegisterHandler,
	commands.NewRoomsAliasRevokeHandler,

	commands.NewProcessRoomAttendantEventHandler,
	wire.Bind(new(pubsub.ProcessRoomAttendantEventHandler), new(*commands.ProcessRoomAttendantEventHandler)),
)

var queriesSet = wire.NewSet(
	wire.Struct(new(app.Queries), "*"),

	queries.NewReceiveLogHandler,
	queries.NewPublishedLogHandler,
	queries.NewStatusHandler,
	queries.NewBlobDownloadedEventsHandler,
	queries.NewRoomsListAliasesHandler,
	queries.NewGetMessageBySequenceHandler,

	queries.NewCreateHistoryStreamHandler,
	wire.Bind(new(portsrpc.CreateHistoryStreamQueryHandler), new(*queries.CreateHistoryStreamHandler)),
	wire.Bind(new(ebtadapters.CreateHistoryStreamHandler), new(*queries.CreateHistoryStreamHandler)),

	queries.NewGetBlobHandler,
	wire.Bind(new(portsrpc.GetBlobQueryHandler), new(*queries.GetBlobHandler)),
)
