//go:build wireinject
// +build wireinject

package di

import (
	"context"
	"path"
	"time"

	"github.com/boreq/errors"
	"github.com/google/wire"
	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/adapters/bolt"
	"github.com/planetary-social/go-ssb/service/adapters/mocks"
	"github.com/planetary-social/go-ssb/service/adapters/pubsub"
	"github.com/planetary-social/go-ssb/service/app"
	"github.com/planetary-social/go-ssb/service/app/commands"
	"github.com/planetary-social/go-ssb/service/app/queries"
	"github.com/planetary-social/go-ssb/service/domain"
	"github.com/planetary-social/go-ssb/service/domain/feeds"
	"github.com/planetary-social/go-ssb/service/domain/feeds/content/transport"
	"github.com/planetary-social/go-ssb/service/domain/feeds/formats"
	"github.com/planetary-social/go-ssb/service/domain/graph"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/network"
	"github.com/planetary-social/go-ssb/service/domain/network/local"
	"github.com/planetary-social/go-ssb/service/domain/replication"
	domaintransport "github.com/planetary-social/go-ssb/service/domain/transport"
	"github.com/planetary-social/go-ssb/service/domain/transport/boxstream"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
	portsnetwork "github.com/planetary-social/go-ssb/service/ports/network"
	portspubsub "github.com/planetary-social/go-ssb/service/ports/pubsub"
	portsrpc "github.com/planetary-social/go-ssb/service/ports/rpc"
	"go.etcd.io/bbolt"
)

var replicatorSet = wire.NewSet(
	replication.NewManager,
	wire.Bind(new(replication.ReplicationManager), new(*replication.Manager)),

	replication.NewGossipReplicator,
	wire.Bind(new(domain.Replicator), new(*replication.GossipReplicator)),
)

var formatsSet = wire.NewSet(
	newFormats,

	formats.NewScuttlebutt,

	transport.NewMarshaler,
	wire.Bind(new(formats.Marshaler), new(*transport.Marshaler)),

	transport.DefaultMappings,

	formats.NewRawMessageIdentifier,
	wire.Bind(new(commands.RawMessageIdentifier), new(*formats.RawMessageIdentifier)),
	wire.Bind(new(bolt.RawMessageIdentifier), new(*formats.RawMessageIdentifier)),
)

var portsSet = wire.NewSet(
	portsrpc.NewMux,

	portsrpc.NewMuxHandlers,
	portsrpc.NewHandlerBlobsGet,
	portsrpc.NewHandlerCreateHistoryStream,

	portspubsub.NewPubSub,

	local.NewDiscoverer,
	portsnetwork.NewDiscoverer,
	portsnetwork.NewConnectionEstablisher,
)

var requestPubSubSet = wire.NewSet(
	pubsub.NewRequestPubSub,
	wire.Bind(new(rpc.RequestHandler), new(*pubsub.RequestPubSub)),
)

var messagePubSubSet = wire.NewSet(
	pubsub.NewMessagePubSub,
	wire.Bind(new(queries.MessageSubscriber), new(*pubsub.MessagePubSub)),
)

var hops = graph.MustNewHops(3)

type TxTestAdapters struct {
	MessageRepository *bolt.MessageRepository
	FeedRepository    *bolt.FeedRepository
	ReceiveLog        *bolt.ReceiveLogRepository
}

func BuildTxAdaptersForTest(*bbolt.Tx) (TxTestAdapters, error) {
	wire.Build(
		wire.Struct(new(TxTestAdapters), "*"),

		txBoltAdaptersSet,

		identity.NewPrivate,
		privateIdentityToPublicIdentity,

		formats.NewDefaultMessageHMAC,

		fixtures.SomeLogger,

		formatsSet,
		wire.Value(hops),
	)

	return TxTestAdapters{}, nil
}

type TestAdapters struct {
	MessageRepository *bolt.ReadMessageRepository
	FeedRepository    *bolt.ReadFeedRepository
	ReceiveLog        *bolt.ReadReceiveLogRepository
}

func BuildAdaptersForTest(*bbolt.DB) (TestAdapters, error) {
	wire.Build(
		wire.Struct(new(TestAdapters), "*"),

		boltAdaptersSet,

		identity.NewPrivate,
		privateIdentityToPublicIdentity,

		formats.NewDefaultMessageHMAC,

		fixtures.SomeLogger,
	)

	return TestAdapters{}, nil
}

type TestApplication struct {
	Queries app.Queries

	FeedRepository    *mocks.FeedRepositoryMock
	MessagePubSub     *mocks.MessagePubSubMock
	MessageRepository *mocks.MessageRepositoryMock
	PeerManager       *mocks.PeerManagerMock
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

		mocks.NewPeerManagerMock,
		wire.Bind(new(queries.PeerManager), new(*mocks.PeerManagerMock)),

		wire.Struct(new(TestApplication), "*"),

		identity.NewPrivate,
		privateIdentityToPublicIdentity,
	)

	return TestApplication{}, nil

}

func BuildTransactableAdapters(*bbolt.Tx, identity.Public, logging.Logger, Config) (commands.Adapters, error) {
	wire.Build(
		wire.Struct(new(commands.Adapters), "*"),

		txBoltAdaptersSet,
		formatsSet,

		extractMessageHMACFromConfig,

		wire.Value(hops),
	)

	return commands.Adapters{}, nil
}

func BuildTxRepositories(*bbolt.Tx, identity.Public, logging.Logger, formats.MessageHMAC) (bolt.TxRepositories, error) {
	wire.Build(
		wire.Struct(new(bolt.TxRepositories), "*"),

		txBoltAdaptersSet,
		formatsSet,

		wire.Value(hops),
	)

	return bolt.TxRepositories{}, nil
}

// BuildService creates a new service which uses the provided context as a long-term context used as a base context for
// e.g. established connections.
func BuildService(context.Context, identity.Private, Config) (Service, error) {
	wire.Build(
		NewService,

		extractNetworkKeyFromConfig,
		extractMessageHMACFromConfig,
		extractLoggerFromConfig,
		extractPeerManagerConfigFromConfig,

		boxstream.NewHandshaker,

		commands.NewRawMessageHandler,
		wire.Bind(new(replication.RawMessageHandler), new(*commands.RawMessageHandler)),

		domaintransport.NewPeerInitializer,
		wire.Bind(new(portsnetwork.ServerPeerInitializer), new(*domaintransport.PeerInitializer)),
		wire.Bind(new(network.ClientPeerInitializer), new(*domaintransport.PeerInitializer)),

		network.NewDialer,
		wire.Bind(new(commands.Dialer), new(*network.Dialer)),
		wire.Bind(new(domain.Dialer), new(*network.Dialer)),

		domain.NewPeerManager,
		wire.Bind(new(commands.NewPeerHandler), new(*domain.PeerManager)),
		wire.Bind(new(commands.PeerManager), new(*domain.PeerManager)),
		wire.Bind(new(queries.PeerManager), new(*domain.PeerManager)),

		bolt.NewTransactionProvider,
		wire.Bind(new(commands.TransactionProvider), new(*bolt.TransactionProvider)),
		newAdaptersFactory,

		newAdvertiser,
		newListener,
		privateIdentityToPublicIdentity,

		commands.NewMessageBuffer,

		portsSet,
		applicationSet,
		replicatorSet,
		formatsSet,
		requestPubSubSet,
		messagePubSubSet,
		boltAdaptersSet,

		newBolt,
	)
	return Service{}, nil
}

func newAdvertiser(l identity.Public, config Config) (*local.Advertiser, error) {
	return local.NewAdvertiser(l, config.ListenAddress)
}

func newListener(
	ctx context.Context,
	initializer portsnetwork.ServerPeerInitializer,
	app app.Application,
	config Config,
	logger logging.Logger,
) (*portsnetwork.Listener, error) {
	return portsnetwork.NewListener(ctx, initializer, app, config.ListenAddress, logger)
}

func newAdaptersFactory(config Config, local identity.Public, logger logging.Logger) bolt.AdaptersFactory {
	return func(tx *bbolt.Tx) (commands.Adapters, error) {
		return BuildTransactableAdapters(tx, local, logger, config)
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

func newFormats(
	s *formats.Scuttlebutt,
) []feeds.FeedFormat {
	return []feeds.FeedFormat{
		s,
	}
}

func extractNetworkKeyFromConfig(config Config) boxstream.NetworkKey {
	return config.NetworkKey
}

func extractMessageHMACFromConfig(config Config) formats.MessageHMAC {
	return config.MessageHMAC
}

func extractLoggerFromConfig(config Config) logging.Logger {
	return config.Logger
}

func extractPeerManagerConfigFromConfig(config Config) domain.PeerManagerConfig {
	return config.PeerManagerConfig
}
