//go:build wireinject
// +build wireinject

package di

import (
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
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/content/transport"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/formats"
	"github.com/planetary-social/go-ssb/scuttlebutt/replication"
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
	"time"
)

var applicationSet = wire.NewSet(
	wire.Struct(new(commands.Application), "*"),
	commands.NewRedeemInviteHandler,
	commands.NewFollowHandler,
)

func BuildService(identity.Private) (Service, error) {
	wire.Build(
		NewService,

		boxstream.NewDefaultNetworkKey,

		boxstream.NewHandshaker,

		network.NewListener,

		replication.NewManager,
		wire.Bind(new(replication.ReplicationManager), new(*replication.Manager)),

		commands.NewRawMessageHandler,
		wire.Bind(new(replication.RawMessageHandler), new(*commands.RawMessageHandler)),

		adapters.NewBoltFeedStorage,
		wire.Bind(new(commands.FeedStorage), new(*adapters.BoltFeedStorage)),
		wire.Bind(new(replication.FeedStorage), new(*adapters.BoltFeedStorage)),

		network.NewPeerInitializer,
		wire.Bind(new(network.ServerPeerInitializer), new(*network.PeerInitializer)),
		wire.Bind(new(network.ClientPeerInitializer), new(*network.PeerInitializer)),

		network.NewDialer,
		wire.Bind(new(commands.Dialer), new(*network.Dialer)),

		rpc.NewMux,
		wire.Bind(new(rpc.RequestHandler), new(*rpc.Mux)),

		replication.NewGossipReplicator,
		wire.Bind(new(scuttlebutt.Replicator), new(*replication.GossipReplicator)),

		scuttlebutt.NewPeerManager,
		wire.Bind(new(network.NewPeerHandler), new(*scuttlebutt.PeerManager)),

		formats.NewRawMessageIdentifier,
		wire.Bind(new(commands.RawMessageIdentifier), new(*formats.RawMessageIdentifier)),
		wire.Bind(new(adapters.RawMessageIdentifier), new(*formats.RawMessageIdentifier)),

		formats.AllFormats,

		transport.NewMarshaler,
		wire.Bind(new(formats.Marshaler), new(*transport.Marshaler)),

		transport.DefaultMappings,

		applicationSet,

		newLogger,

		newBolt,
	)
	return Service{}, nil
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
