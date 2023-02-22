package di

import (
	"github.com/google/wire"
	ebtadapters "github.com/planetary-social/scuttlego/service/adapters/ebt"
	"github.com/planetary-social/scuttlego/service/app"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/ports/network"
	"github.com/planetary-social/scuttlego/service/ports/pubsub"
	portsrpc "github.com/planetary-social/scuttlego/service/ports/rpc"
)

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
	commands.NewPublishRawHandler,
	commands.NewPublishRawAsIdentityHandler,
	commands.NewDownloadBlobHandler,
	commands.NewCreateBlobHandler,
	commands.NewDownloadFeedHandler,
	commands.NewRoomsAliasRegisterHandler,
	commands.NewRoomsAliasRevokeHandler,
	commands.NewAddToBanListHandler,
	commands.NewRemoveFromBanListHandler,
	commands.NewRunMigrationsHandler,

	commands.NewProcessNewLocalDiscoveryHandler,
	wire.Bind(new(network.ProcessNewLocalDiscoveryCommandHandler), new(*commands.ProcessNewLocalDiscoveryHandler)),

	commands.NewAcceptNewPeerHandler,
	wire.Bind(new(pubsub.AcceptNewPeerCommandHandler), new(*commands.AcceptNewPeerHandler)),

	commands.NewEstablishNewConnectionsHandler,
	wire.Bind(new(network.EstablishNewConnectionsCommandHandler), new(*commands.EstablishNewConnectionsHandler)),

	commands.NewRawMessageHandler,
	wire.Bind(new(replication.RawMessageHandler), new(*commands.RawMessageHandler)),

	commands.NewCreateWantsHandler,
	wire.Bind(new(portsrpc.CreateWantsCommandHandler), new(*commands.CreateWantsHandler)),

	commands.NewHandleIncomingEbtReplicateHandler,
	wire.Bind(new(portsrpc.EbtReplicateCommandHandler), new(*commands.HandleIncomingEbtReplicateHandler)),

	commands.NewProcessRoomAttendantEventHandler,
	wire.Bind(new(pubsub.ProcessRoomAttendantEventHandler), new(*commands.ProcessRoomAttendantEventHandler)),

	commands.NewAcceptTunnelConnectHandler,
	wire.Bind(new(portsrpc.AcceptTunnelConnectHandler), new(*commands.AcceptTunnelConnectHandler)),
)

var queriesSet = wire.NewSet(
	wire.Struct(new(app.Queries), "*"),

	queries.NewReceiveLogHandler,
	queries.NewPublishedLogHandler,
	queries.NewStatusHandler,
	queries.NewBlobDownloadedEventsHandler,
	queries.NewRoomsListAliasesHandler,
	queries.NewGetMessageHandler,
	queries.NewGetMessageBySequenceHandler,

	queries.NewCreateHistoryStreamHandler,
	wire.Bind(new(portsrpc.CreateHistoryStreamQueryHandler), new(*queries.CreateHistoryStreamHandler)),
	wire.Bind(new(ebtadapters.CreateHistoryStreamHandler), new(*queries.CreateHistoryStreamHandler)),

	queries.NewGetBlobHandler,
	wire.Bind(new(portsrpc.GetBlobQueryHandler), new(*queries.GetBlobHandler)),
)
