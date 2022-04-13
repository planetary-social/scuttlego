//go:build wireinject
// +build wireinject

package di

import (
	"path"
	"time"

	"github.com/boreq/errors"
	"github.com/google/wire"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/adapters/bolt"
	"github.com/planetary-social/go-ssb/service/adapters/mocks"
	"github.com/planetary-social/go-ssb/service/adapters/pubsub"
	"github.com/planetary-social/go-ssb/service/app"
	commands2 "github.com/planetary-social/go-ssb/service/app/commands"
	"github.com/planetary-social/go-ssb/service/app/queries"
	"github.com/planetary-social/go-ssb/service/domain"
	"github.com/planetary-social/go-ssb/service/domain/feeds"
	transport2 "github.com/planetary-social/go-ssb/service/domain/feeds/content/transport"
	formats2 "github.com/planetary-social/go-ssb/service/domain/feeds/formats"
	"github.com/planetary-social/go-ssb/service/domain/graph"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/network"
	"github.com/planetary-social/go-ssb/service/domain/network/local"
	replication2 "github.com/planetary-social/go-ssb/service/domain/replication"
	network2 "github.com/planetary-social/go-ssb/service/domain/transport"
	boxstream2 "github.com/planetary-social/go-ssb/service/domain/transport/boxstream"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
	portsnetwork "github.com/planetary-social/go-ssb/service/ports/network"
	portspubsub "github.com/planetary-social/go-ssb/service/ports/pubsub"
	portsrpc "github.com/planetary-social/go-ssb/service/ports/rpc"
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
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
	wire.Bind(new(bolt.RawMessageIdentifier), new(*formats2.RawMessageIdentifier)),
)

var portsSet = wire.NewSet(
	portsrpc.NewMux,

	portsrpc.NewMuxHandlers,
	portsrpc.NewHandlerBlobsGet,
	portsrpc.NewHandlerCreateHistoryStream,

	portspubsub.NewPubSub,

	local.NewDiscoverer,
	portsnetwork.NewDiscoverer,
)

var requestPubSubSet = wire.NewSet(
	pubsub.NewRequestPubSub,
	wire.Bind(new(rpc.RequestHandler), new(*pubsub.RequestPubSub)),
)

var messagePubSubSet = wire.NewSet(
	pubsub.NewMessagePubSub,
	wire.Bind(new(queries.MessageSubscriber), new(*pubsub.MessagePubSub)),
)

var adaptersSet = wire.NewSet(
	bolt.NewBoltFeedMessagesRepository,
	wire.Bind(new(queries.FeedRepository), new(*bolt.BoltFeedMessagesRepository)),
)

type TestAdapters struct {
	Feed *bolt.FeedRepository
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

type TestApplication struct {
	Queries app.Queries

	FeedRepository    *mocks.FeedRepositoryMock
	MessagePubSub     *mocks.MessagePubSubMock
	MessageRepository *mocks.MessageRepositoryMock
}

func BuildApplicationForTests() (TestApplication, error) {
	wire.Build(
		applicationSet,

		mocks.NewMessagePubSubMock,
		wire.Bind(new(queries.MessageSubscriber), new(*mocks.MessagePubSubMock)),

		mocks.NewFeedRepositoryMock,
		wire.Bind(new(queries.FeedRepository), new(*mocks.FeedRepositoryMock)),

		mocks.NewReceiveLogRepositoryMock,
		wire.Bind(new(queries.ReceiveLogRepository), new(*mocks.ReceiveLogRepositoryMock)),

		mocks.NewMessageRepositoryMock,
		wire.Bind(new(queries.MessageRepository), new(*mocks.MessageRepositoryMock)),

		wire.Struct(new(TestApplication), "*"),
	)

	return TestApplication{}, nil

}

func BuildTransactableAdapters(*bbolt.Tx, identity.Private, logging.Logger, Config) (commands2.Adapters, error) {
	wire.Build(
		wire.Struct(new(commands2.Adapters), "*"),

		bolt.NewFeedRepository,
		wire.Bind(new(commands2.FeedRepository), new(*bolt.FeedRepository)),

		bolt.NewSocialGraphRepository,
		wire.Bind(new(commands2.SocialGraphRepository), new(*bolt.SocialGraphRepository)),

		bolt.NewReceiveLogRepository,
		bolt.NewMessageRepository,

		formatsSet,

		privateIdentityToPublicIdentity,

		newMessageHMACFromConfig,

		wire.Value(hops),
	)

	return commands2.Adapters{}, nil
}

func BuildAdaptersForContactsRepository(*bbolt.Tx, identity.Private, logging.Logger, Config) (bolt.Repositories, error) {
	wire.Build(
		wire.Struct(new(bolt.Repositories), "*"),

		bolt.NewFeedRepository,
		bolt.NewSocialGraphRepository,
		bolt.NewReceiveLogRepository,
		bolt.NewMessageRepository,

		formatsSet,

		privateIdentityToPublicIdentity,

		newMessageHMACFromConfig,

		wire.Value(hops),
	)

	return bolt.Repositories{}, nil
}

func BuildService(identity.Private, Config) (Service, error) {
	wire.Build(
		NewService,

		newNetworkKeyFromConfig,
		newMessageHMACFromConfig,

		boxstream2.NewHandshaker,

		commands2.NewRawMessageHandler,
		wire.Bind(new(replication2.RawMessageHandler), new(*commands2.RawMessageHandler)),

		network2.NewPeerInitializer,
		wire.Bind(new(portsnetwork.ServerPeerInitializer), new(*network2.PeerInitializer)),
		wire.Bind(new(network.ClientPeerInitializer), new(*network2.PeerInitializer)),

		network.NewDialer,
		wire.Bind(new(commands2.Dialer), new(*network.Dialer)),

		domain.NewPeerManager,
		wire.Bind(new(commands2.NewPeerHandler), new(*domain.PeerManager)),

		bolt.NewTransactionProvider,
		wire.Bind(new(commands2.TransactionProvider), new(*bolt.TransactionProvider)),
		newAdaptersFactory,

		bolt.NewBoltContactsRepository,
		wire.Bind(new(replication2.Storage), new(*bolt.BoltContactsRepository)),
		newContactRepositoriesFactory,

		bolt.NewReceiveLogReadRepository,
		wire.Bind(new(queries.ReceiveLogRepository), new(*bolt.ReceiveLogReadRepository)),

		bolt.NewReadMessageRepository,
		wire.Bind(new(queries.MessageRepository), new(*bolt.ReadMessageRepository)),

		newAdvertiser,
		newListener,
		privateIdentityToPublicIdentity,

		portsSet,
		applicationSet,
		replicatorSet,
		formatsSet,
		requestPubSubSet,
		messagePubSubSet,
		adaptersSet,

		newLogger,

		newBolt,
	)
	return Service{}, nil
}

func newAdvertiser(l identity.Public, config Config) (*local.Advertiser, error) {
	return local.NewAdvertiser(l, config.ListenAddress)
}

func newListener(
	initializer portsnetwork.ServerPeerInitializer,
	app app.Application,
	config Config,
	logger logging.Logger,
) (*portsnetwork.Listener, error) {
	return portsnetwork.NewListener(initializer, app, config.ListenAddress, logger)
}

func newAdaptersFactory(config Config, local identity.Private, logger logging.Logger) bolt.AdaptersFactory {
	return func(tx *bbolt.Tx) (commands2.Adapters, error) {
		return BuildTransactableAdapters(tx, local, logger, config)
	}
}

func newContactRepositoriesFactory(local identity.Private, logger logging.Logger, config Config) bolt.RepositoriesFactory {
	return func(tx *bbolt.Tx) (bolt.Repositories, error) {
		return BuildAdaptersForContactsRepository(tx, local, logger, config)
	}
}

func newBolt(config Config) (*bbolt.DB, error) {
	filename := path.Join(config.DataDirectory, "database.bolt")
	b, err := bbolt.Open(filename, 0600, &bbolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, errors.Wrap(err, "could not open the database, is something else reading it?")
	}
	return b, nil
}

func privateIdentityToPublicIdentity(p identity.Private) identity.Public {
	return p.Public()
}

func newLogger(config Config) logging.Logger {
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	return logging.NewLogrusLogger(log, "main", config.LoggingLevel)
}

func newFormats(
	s *formats2.Scuttlebutt,
) []feeds.FeedFormat {
	return []feeds.FeedFormat{
		s,
	}
}

func newNetworkKeyFromConfig(config Config) boxstream2.NetworkKey {
	return config.NetworkKey
}

func newMessageHMACFromConfig(config Config) formats2.MessageHMAC {
	return config.MessageHMAC
}
