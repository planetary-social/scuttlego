//go:build wireinject
// +build wireinject

package di

import (
	rpcHandlers "github.com/planetary-social/go-ssb/scuttlebutt/rpc"
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
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/content/transport"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/formats"
	"github.com/planetary-social/go-ssb/scuttlebutt/graph"
	"github.com/planetary-social/go-ssb/scuttlebutt/replication"
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
)

var applicationSet = wire.NewSet(
	wire.Struct(new(commands.Application), "*"),
	commands.NewRedeemInviteHandler,
	commands.NewFollowHandler,
	commands.NewConnectHandler,
)

var replicatorSet = wire.NewSet(
	replication.NewManager,
	wire.Bind(new(replication.ReplicationManager), new(*replication.Manager)),

	replication.NewGossipReplicator,
	wire.Bind(new(scuttlebutt.Replicator), new(*replication.GossipReplicator)),
)

var formatsSet = wire.NewSet(
	newFormats,

	formats.NewScuttlebutt,

	transport.NewMarshaler,
	wire.Bind(new(formats.Marshaler), new(*transport.Marshaler)),

	transport.DefaultMappings,

	formats.NewRawMessageIdentifier,
	wire.Bind(new(commands.RawMessageIdentifier), new(*formats.RawMessageIdentifier)),
	wire.Bind(new(adapters.RawMessageIdentifier), new(*formats.RawMessageIdentifier)),
)

var muxSet = wire.NewSet(
	newMuxWithHandlers,
	wire.Bind(new(rpc.RequestHandler), new(*rpc.Mux)),

	newMuxHandlers,
	rpcHandlers.NewHandlerBlobsGet,
	rpcHandlers.NewHandlerCreateHistoryStream,
)

type TestAdapters struct {
	Feed *adapters.BoltFeedRepository
}

//func BuildAdaptersForTest(*bbolt.Tx) (TestAdapters, error) {
//	wire.Build(
//		wire.Struct(new(TestAdapters), "*"),
//
//		adapters.NewBoltFeedRepository,
//		adapters.NewSocialGraphRepository,
//
//		newLogger,
//
//		formats.NewRawMessageIdentifier,
//		wire.Bind(new(adapters.RawMessageIdentifier), new(*formats.RawMessageIdentifier)),
//
//		formatsSet,
//	)
//
//	return TestAdapters{}, nil
//
//}

var hops = graph.MustNewHops(3)

func BuildAdapters(*bbolt.Tx, identity.Private) (commands.Adapters, error) {
	wire.Build(
		wire.Struct(new(commands.Adapters), "*"),

		adapters.NewBoltFeedRepository,
		wire.Bind(new(commands.FeedRepository), new(*adapters.BoltFeedRepository)),

		adapters.NewSocialGraphRepository,
		wire.Bind(new(commands.SocialGraphRepository), new(*adapters.SocialGraphRepository)),

		formatsSet,

		newPublicIdentity,
		newLogger,

		wire.Value(hops),
	)

	return commands.Adapters{}, nil
}

func BuildAdaptersForContactsRepository(*bbolt.Tx, identity.Private) (adapters.Repositories, error) {
	wire.Build(
		wire.Struct(new(adapters.Repositories), "*"),

		adapters.NewBoltFeedRepository,
		adapters.NewSocialGraphRepository,

		formatsSet,

		newPublicIdentity,
		newLogger,

		wire.Value(hops),
	)

	return adapters.Repositories{}, nil
}

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

		scuttlebutt.NewPeerManager,
		wire.Bind(new(network.NewPeerHandler), new(*scuttlebutt.PeerManager)),
		wire.Bind(new(commands.NewPeerHandler), new(*scuttlebutt.PeerManager)),

		adapters.NewTransactionProvider,
		wire.Bind(new(commands.TransactionProvider), new(*adapters.TransactionProvider)),
		newAdaptersFactory,

		//wire.Bind(new(adapters.AdaptersFactory), new(BuildAdapters))
		adapters.NewBoltContactsRepository,
		wire.Bind(new(replication.Storage), new(*adapters.BoltContactsRepository)),
		newContactRepositoriesFactory,

		muxSet,
		applicationSet,
		replicatorSet,
		formatsSet,

		newLogger,

		newBolt,
	)
	return Service{}, nil
}

func newMuxHandlers(
	createHistoryStream *rpcHandlers.HandlerCreateHistoryStream,
	blobsGet *rpcHandlers.HandlerBlobsGet,
) []rpc.Handler {
	return []rpc.Handler{
		createHistoryStream,
		blobsGet,
	}
}

func newMuxWithHandlers(logger logging.Logger, handlers []rpc.Handler) (*rpc.Mux, error) {
	mux := rpc.NewMux(logger)
	for _, handler := range handlers {
		if err := mux.AddHandler(handler); err != nil {
			return nil, err
		}
	}
	return mux, nil
}

func newAdaptersFactory(local identity.Private) adapters.AdaptersFactory {
	return func(tx *bbolt.Tx) (commands.Adapters, error) {
		return BuildAdapters(tx, local)
	}
}

func newContactRepositoriesFactory(local identity.Private) adapters.RepositoriesFactory {
	return func(tx *bbolt.Tx) (adapters.Repositories, error) {
		return BuildAdaptersForContactsRepository(tx, local)
	}
}

func newBolt() (*bbolt.DB, error) {
	b, err := bbolt.Open("/tmp/tmp.bolt.db", 0600, &bbolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, errors.Wrap(err, "could not open the database, is something else reading it?")
	}
	return b, nil
}

func newPublicIdentity(p identity.Private) identity.Public {
	return p.Public()
}

func newLogger() logging.Logger {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	return logging.NewLogrusLogger(log, "main", logging.LevelDebug)
}

func newFormats(
	s *formats.Scuttlebutt,
) []feeds.FeedFormat {
	return []feeds.FeedFormat{
		s,
	}
}
