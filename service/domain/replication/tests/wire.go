//go:build wireinject
// +build wireinject

package tests

import (
	"testing"

	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/planetary-social/scuttlego/service/domain/replication/gossip"
)

type TestReplication struct {
	Negotiator *replication.Negotiator

	RawMessageHandler  *RawMessageHandlerMock
	ContactsRepository *WantedFeedsRepositoryMock
	MessageStreamer    *MessageStreamerMock
}

func BuildTestReplication(t *testing.T) (TestReplication, error) {
	wire.Build(
		wire.Struct(new(TestReplication), "*"),

		replication.NewNegotiator,

		ebt.NewReplicator,
		wire.Bind(new(replication.EpidemicBroadcastTreesReplicator), new(ebt.Replicator)),

		gossip.NewGossipReplicator,
		wire.Bind(new(replication.CreateHistoryStreamReplicator), new(*gossip.GossipReplicator)),
		wire.Bind(new(ebt.SelfCreateHistoryStreamReplicator), new(*gossip.GossipReplicator)),

		ebt.NewSessionTracker,
		wire.Bind(new(ebt.Tracker), new(*ebt.SessionTracker)),

		ebt.NewSessionRunner,
		wire.Bind(new(ebt.Runner), new(*ebt.SessionRunner)),

		replication.NewWantedFeedsCache,
		wire.Bind(new(ebt.ContactsStorage), new(*replication.WantedFeedsCache)),
		wire.Bind(new(gossip.ContactsStorage), new(*replication.WantedFeedsCache)),

		gossip.NewManager,
		wire.Bind(new(gossip.ReplicationManager), new(*gossip.Manager)),

		NewRawMessageHandlerMock,
		wire.Bind(new(replication.RawMessageHandler), new(*RawMessageHandlerMock)),

		NewWantedFeedsRepositoryMock,
		wire.Bind(new(replication.WantedFeedsRepository), new(*WantedFeedsRepositoryMock)),

		NewMessageStreamerMock,
		wire.Bind(new(ebt.MessageStreamer), new(*MessageStreamerMock)),

		logging.NewDevNullLogger,
		wire.Bind(new(logging.Logger), new(logging.DevNullLogger)),
	)

	return TestReplication{}, nil
}
