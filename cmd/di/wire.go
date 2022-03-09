//go:build wireinject
// +build wireinject

package di

import (
	adapters2 "github.com/planetary-social/go-ssb/service/adapters"
	"github.com/planetary-social/go-ssb/service/app"
	commands2 "github.com/planetary-social/go-ssb/service/app/commands"
	"github.com/planetary-social/go-ssb/service/domain"
	"github.com/planetary-social/go-ssb/service/domain/feeds"
	transport2 "github.com/planetary-social/go-ssb/service/domain/feeds/content/transport"
	formats2 "github.com/planetary-social/go-ssb/service/domain/feeds/formats"
	"github.com/planetary-social/go-ssb/service/domain/graph"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	network2 "github.com/planetary-social/go-ssb/service/domain/network"
	boxstream2 "github.com/planetary-social/go-ssb/service/domain/network/boxstream"
	rpc3 "github.com/planetary-social/go-ssb/service/domain/network/rpc"
	replication2 "github.com/planetary-social/go-ssb/service/domain/replication"
	rpc2 "github.com/planetary-social/go-ssb/service/ports/rpc"
	"time"

	"github.com/boreq/errors"
	"github.com/google/wire"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
)

var applicationSet = wire.NewSet(
	wire.Struct(new(app.Application), "*"),
	commands2.NewRedeemInviteHandler,
	commands2.NewFollowHandler,
	commands2.NewConnectHandler,
)

var replicatorSet = wire.NewSet(
	replication2.NewManager,
	wire.Bind(new(replication2.ReplicationManager), new(*replication2.Manager)),

	replication2.NewGossipReplicator,
	wire.Bind(new(domain.Replicator), new(*replication2.GossipReplicator)),
)

var formatsSet = wire.NewSet(
	newFormats,

	formats2.NewScuttlebutt,

	transport2.NewMarshaler,
	wire.Bind(new(formats2.Marshaler), new(*transport2.Marshaler)),

	transport2.DefaultMappings,

	formats2.NewRawMessageIdentifier,
	wire.Bind(new(commands2.RawMessageIdentifier), new(*formats2.RawMessageIdentifier)),
	wire.Bind(new(adapters2.RawMessageIdentifier), new(*formats2.RawMessageIdentifier)),
)

var muxSet = wire.NewSet(
	newMuxWithHandlers,
	wire.Bind(new(rpc3.RequestHandler), new(*rpc3.Mux)),

	newMuxHandlers,
	rpc2.NewHandlerBlobsGet,
	rpc2.NewHandlerCreateHistoryStream,
)

type TestAdapters struct {
	Feed *adapters2.BoltFeedRepository
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

func BuildAdapters(*bbolt.Tx, identity.Private) (commands2.Adapters, error) {
	wire.Build(
		wire.Struct(new(commands2.Adapters), "*"),

		adapters2.NewBoltFeedRepository,
		wire.Bind(new(commands2.FeedRepository), new(*adapters2.BoltFeedRepository)),

		adapters2.NewSocialGraphRepository,
		wire.Bind(new(commands2.SocialGraphRepository), new(*adapters2.SocialGraphRepository)),

		formatsSet,

		newPublicIdentity,
		newLogger,

		wire.Value(hops),
	)

	return commands2.Adapters{}, nil
}

func BuildAdaptersForContactsRepository(*bbolt.Tx, identity.Private) (adapters2.Repositories, error) {
	wire.Build(
		wire.Struct(new(adapters2.Repositories), "*"),

		adapters2.NewBoltFeedRepository,
		adapters2.NewSocialGraphRepository,

		formatsSet,

		newPublicIdentity,
		newLogger,

		wire.Value(hops),
	)

	return adapters2.Repositories{}, nil
}

func BuildService(identity.Private) (Service, error) {
	wire.Build(
		NewService,

		boxstream2.NewDefaultNetworkKey,

		boxstream2.NewHandshaker,

		network2.NewListener,

		commands2.NewRawMessageHandler,
		wire.Bind(new(replication2.RawMessageHandler), new(*commands2.RawMessageHandler)),

		network2.NewPeerInitializer,
		wire.Bind(new(network2.ServerPeerInitializer), new(*network2.PeerInitializer)),
		wire.Bind(new(network2.ClientPeerInitializer), new(*network2.PeerInitializer)),

		network2.NewDialer,
		wire.Bind(new(commands2.Dialer), new(*network2.Dialer)),

		domain.NewPeerManager,
		wire.Bind(new(network2.NewPeerHandler), new(*domain.PeerManager)),
		wire.Bind(new(commands2.NewPeerHandler), new(*domain.PeerManager)),

		adapters2.NewTransactionProvider,
		wire.Bind(new(commands2.TransactionProvider), new(*adapters2.TransactionProvider)),
		newAdaptersFactory,

		//wire.Bind(new(adapters.AdaptersFactory), new(BuildAdapters))
		adapters2.NewBoltContactsRepository,
		wire.Bind(new(replication2.Storage), new(*adapters2.BoltContactsRepository)),
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
	createHistoryStream *rpc2.HandlerCreateHistoryStream,
	blobsGet *rpc2.HandlerBlobsGet,
) []rpc3.Handler {
	return []rpc3.Handler{
		createHistoryStream,
		blobsGet,
	}
}

func newMuxWithHandlers(logger logging.Logger, handlers []rpc3.Handler) (*rpc3.Mux, error) {
	mux := rpc3.NewMux(logger)
	for _, handler := range handlers {
		if err := mux.AddHandler(handler); err != nil {
			return nil, err
		}
	}
	return mux, nil
}

func newAdaptersFactory(local identity.Private) adapters2.AdaptersFactory {
	return func(tx *bbolt.Tx) (commands2.Adapters, error) {
		return BuildAdapters(tx, local)
	}
}

func newContactRepositoriesFactory(local identity.Private) adapters2.RepositoriesFactory {
	return func(tx *bbolt.Tx) (adapters2.Repositories, error) {
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
	s *formats2.Scuttlebutt,
) []feeds.FeedFormat {
	return []feeds.FeedFormat{
		s,
	}
}
