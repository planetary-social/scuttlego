package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/planetary-social/scuttlego/service/domain/replication/gossip"
)

var replicationSet = wire.NewSet(
	gossip.NewManager,
	wire.Bind(new(gossip.ReplicationManager), new(*gossip.Manager)),

	gossip.NewGossipReplicator,
	wire.Bind(new(replication.CreateHistoryStreamReplicator), new(*gossip.GossipReplicator)),
	wire.Bind(new(ebt.SelfCreateHistoryStreamReplicator), new(*gossip.GossipReplicator)),

	ebt.NewReplicator,
	wire.Bind(new(replication.EpidemicBroadcastTreesReplicator), new(ebt.Replicator)),

	queries.NewWantedFeedsProvider,
	wire.Bind(new(replication.WantedFeedsProvider), new(*queries.WantedFeedsProvider)),

	replication.NewWantedFeedsCache,
	wire.Bind(new(replication.ContactsStorage), new(*replication.WantedFeedsCache)),
	wire.Bind(new(commands.ForkedFeedTracker), new(*replication.WantedFeedsCache)),

	ebt.NewSessionTracker,
	wire.Bind(new(ebt.Tracker), new(*ebt.SessionTracker)),

	ebt.NewSessionRunner,
	wire.Bind(new(ebt.Runner), new(*ebt.SessionRunner)),

	replication.NewNegotiator,
	wire.Bind(new(commands.MessageReplicator), new(*replication.Negotiator)),
)
