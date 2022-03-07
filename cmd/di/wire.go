//go:build wireinject
// +build wireinject

package di

import (
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/content/transport"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/formats"
	"github.com/planetary-social/go-ssb/scuttlebutt/replication"
	"time"

	"github.com/boreq/errors"
	"github.com/google/wire"
	"github.com/planetary-social/go-ssb/identity"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/network"
	"github.com/planetary-social/go-ssb/network/boxstream"
	"github.com/planetary-social/go-ssb/network/rpc"
	"github.com/planetary-social/go-ssb/scuttlebutt"
	"github.com/planetary-social/go-ssb/scuttlebutt/adapters"
	"github.com/planetary-social/go-ssb/scuttlebutt/commands"
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
)

var applicationSet = wire.NewSet(
	wire.Struct(new(commands.Application), "*"),
	commands.NewRedeemInviteHandler,
	commands.NewFollowHandler,
	commands.NewConnectHandler,
)

type TestAdapters struct {
	Feed *adapters.BoltFeedRepository
}

func BuildAdaptersForTest(*bbolt.Tx) (TestAdapters, error) {
	wire.Build(
		wire.Struct(new(TestAdapters), "*"),

		adapters.NewBoltFeedRepository,
		adapters.NewSocialGraphRepository,

		newLogger,

		formats.NewRawMessageIdentifier,
		wire.Bind(new(adapters.RawMessageIdentifier), new(*formats.RawMessageIdentifier)),

		formats.AllFormats,

		transport.NewMarshaler,
		wire.Bind(new(formats.Marshaler), new(*transport.Marshaler)),

		transport.DefaultMappings,
	)

	return TestAdapters{}, nil

}

func BuildAdapters(*bbolt.Tx, adapters.RawMessageIdentifier) (commands.Adapters, error) {
	wire.Build(
		wire.Struct(new(commands.Adapters), "*"),

		adapters.NewBoltFeedRepository,
		wire.Bind(new(commands.FeedRepository), new(*adapters.BoltFeedRepository)),

		adapters.NewSocialGraphRepository,
		wire.Bind(new(commands.SocialGraphRepository), new(*adapters.SocialGraphRepository)),
	)

	return commands.Adapters{}, nil
}

var replicatorSet = wire.NewSet(
	replication.NewManager,
	wire.Bind(new(replication.ReplicationManager), new(*replication.Manager)),

	replication.NewGossipReplicator,
	wire.Bind(new(scuttlebutt.Replicator), new(*replication.GossipReplicator)),
)

func BuildService(identity.Private) (Service, error) {
	wire.Build(
		NewService,

		boxstream.NewDefaultNetworkKey,

		boxstream.NewHandshaker,

		network.NewListener,

		commands.NewRawMessageHandler,
		wire.Bind(new(replication.RawMessageHandler), new(*commands.RawMessageHandler)),

		network.NewPeerInitializer,
		wire.Bind(new(network.ServerPeerInitializer), new(*network.PeerInitializer)),
		wire.Bind(new(network.ClientPeerInitializer), new(*network.PeerInitializer)),

		network.NewDialer,
		wire.Bind(new(commands.Dialer), new(*network.Dialer)),

		rpc.NewMux,
		wire.Bind(new(rpc.RequestHandler), new(*rpc.Mux)),

		scuttlebutt.NewPeerManager,
		wire.Bind(new(network.NewPeerHandler), new(*scuttlebutt.PeerManager)),
		wire.Bind(new(commands.NewPeerHandler), new(*scuttlebutt.PeerManager)),

		formats.NewRawMessageIdentifier,
		wire.Bind(new(commands.RawMessageIdentifier), new(*formats.RawMessageIdentifier)),
		wire.Bind(new(adapters.RawMessageIdentifier), new(*formats.RawMessageIdentifier)),

		formats.AllFormats,

		transport.NewMarshaler,
		wire.Bind(new(formats.Marshaler), new(*transport.Marshaler)),

		transport.DefaultMappings,

		adapters.NewTransactionProvider,
		wire.Bind(new(commands.TransactionProvider), new(*adapters.TransactionProvider)),
		newAdaptersFactory,

		//wire.Bind(new(adapters.AdaptersFactory), new(BuildAdapters))

		applicationSet,
		replicatorSet,

		newLogger,

		newBolt,
	)
	return Service{}, nil
}

func newAdaptersFactory(identifier adapters.RawMessageIdentifier) adapters.AdaptersFactory {
	return func(tx *bbolt.Tx) (commands.Adapters, error) {
		return BuildAdapters(tx, identifier)
	}
}

func newBolt() (*bbolt.DB, error) {
	b, err := bbolt.Open("/tmp/tmp.bolt.db", 0600, &bbolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, errors.Wrap(err, "could not open the database, is something else reading it?")
	}
	return b, nil
}

func newLogger() logging.Logger {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	return logging.NewLogrusLogger(log, "main")
}
