// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/adapters"
	"github.com/planetary-social/scuttlego/service/adapters/bolt"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/adapters/pubsub"
	"github.com/planetary-social/scuttlego/service/app"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain"
	replication2 "github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/transport"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/network/local"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	transport2 "github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux"
	network2 "github.com/planetary-social/scuttlego/service/ports/network"
	pubsub2 "github.com/planetary-social/scuttlego/service/ports/pubsub"
	rpc2 "github.com/planetary-social/scuttlego/service/ports/rpc"
	"go.etcd.io/bbolt"
)

// Injectors from wire.go:

func BuildTxTestAdapters(tx *bbolt.Tx) (TxTestAdapters, error) {
	messageContentMappings := transport.DefaultMappings()
	logger := fixtures.SomeLogger()
	marshaler, err := transport.NewMarshaler(messageContentMappings, logger)
	if err != nil {
		return TxTestAdapters{}, err
	}
	messageHMAC := formats.NewDefaultMessageHMAC()
	scuttlebutt := formats.NewScuttlebutt(marshaler, messageHMAC)
	v := newFormats(scuttlebutt)
	rawMessageIdentifier := formats.NewRawMessageIdentifier(v)
	messageRepository := bolt.NewMessageRepository(tx, rawMessageIdentifier)
	private, err := identity.NewPrivate()
	if err != nil {
		return TxTestAdapters{}, err
	}
	public := privateIdentityToPublicIdentity(private)
	graphHops := _wireHopsValue
	socialGraphRepository := bolt.NewSocialGraphRepository(tx, public, graphHops)
	receiveLogRepository := bolt.NewReceiveLogRepository(tx, messageRepository)
	pubRepository := bolt.NewPubRepository(tx)
	blobRepository := bolt.NewBlobRepository(tx)
	feedRepository := bolt.NewFeedRepository(tx, socialGraphRepository, receiveLogRepository, messageRepository, pubRepository, blobRepository, scuttlebutt)
	currentTimeProviderMock := mocks.NewCurrentTimeProviderMock()
	wantListRepository := bolt.NewWantListRepository(tx, currentTimeProviderMock)
	txTestAdapters := TxTestAdapters{
		MessageRepository:   messageRepository,
		FeedRepository:      feedRepository,
		ReceiveLog:          receiveLogRepository,
		WantList:            wantListRepository,
		CurrentTimeProvider: currentTimeProviderMock,
	}
	return txTestAdapters, nil
}

var (
	_wireHopsValue = hops
)

func BuildTestAdapters(db *bbolt.DB) (TestAdapters, error) {
	private, err := identity.NewPrivate()
	if err != nil {
		return TestAdapters{}, err
	}
	public := privateIdentityToPublicIdentity(private)
	logger := fixtures.SomeLogger()
	messageHMAC := formats.NewDefaultMessageHMAC()
	txRepositoriesFactory := newTxRepositoriesFactory(public, logger, messageHMAC)
	readMessageRepository := bolt.NewReadMessageRepository(db, txRepositoriesFactory)
	readFeedRepository := bolt.NewReadFeedRepository(db, txRepositoriesFactory)
	readReceiveLogRepository := bolt.NewReadReceiveLogRepository(db, txRepositoriesFactory)
	testAdapters := TestAdapters{
		MessageRepository: readMessageRepository,
		FeedRepository:    readFeedRepository,
		ReceiveLog:        readReceiveLogRepository,
	}
	return testAdapters, nil
}

func BuildTestQueries(t *testing.T) (TestQueries, error) {
	feedRepositoryMock := mocks.NewFeedRepositoryMock()
	messagePubSub := pubsub.NewMessagePubSub()
	messagePubSubMock := mocks.NewMessagePubSubMock(messagePubSub)
	logger := fixtures.TestLogger(t)
	createHistoryStreamHandler := queries.NewCreateHistoryStreamHandler(feedRepositoryMock, messagePubSubMock, logger)
	receiveLogRepositoryMock := mocks.NewReceiveLogRepositoryMock()
	receiveLogHandler := queries.NewReceiveLogHandler(receiveLogRepositoryMock)
	private, err := identity.NewPrivate()
	if err != nil {
		return TestQueries{}, err
	}
	public := privateIdentityToPublicIdentity(private)
	publishedLogHandler, err := queries.NewPublishedLogHandler(feedRepositoryMock, receiveLogRepositoryMock, public)
	if err != nil {
		return TestQueries{}, err
	}
	messageRepositoryMock := mocks.NewMessageRepositoryMock()
	peerManagerMock := mocks.NewPeerManagerMock()
	statusHandler := queries.NewStatusHandler(messageRepositoryMock, feedRepositoryMock, peerManagerMock)
	blobStorageMock := mocks.NewBlobStorageMock()
	getBlobHandler, err := queries.NewGetBlobHandler(blobStorageMock)
	if err != nil {
		return TestQueries{}, err
	}
	blobDownloadedPubSubMock := mocks.NewBlobDownloadedPubSubMock()
	blobDownloadedEventsHandler := queries.NewBlobDownloadedEventsHandler(blobDownloadedPubSubMock)
	appQueries := app.Queries{
		CreateHistoryStream:  createHistoryStreamHandler,
		ReceiveLog:           receiveLogHandler,
		PublishedLog:         publishedLogHandler,
		Status:               statusHandler,
		GetBlob:              getBlobHandler,
		BlobDownloadedEvents: blobDownloadedEventsHandler,
	}
	testQueries := TestQueries{
		Queries:              appQueries,
		FeedRepository:       feedRepositoryMock,
		MessagePubSub:        messagePubSubMock,
		MessageRepository:    messageRepositoryMock,
		PeerManager:          peerManagerMock,
		BlobStorage:          blobStorageMock,
		ReceiveLogRepository: receiveLogRepositoryMock,
		LocalIdentity:        public,
	}
	return testQueries, nil
}

func BuildTransactableAdapters(tx *bbolt.Tx, public identity.Public, config Config) (commands.Adapters, error) {
	graphHops := _wireGraphHopsValue
	socialGraphRepository := bolt.NewSocialGraphRepository(tx, public, graphHops)
	messageContentMappings := transport.DefaultMappings()
	logger := extractLoggerFromConfig(config)
	marshaler, err := transport.NewMarshaler(messageContentMappings, logger)
	if err != nil {
		return commands.Adapters{}, err
	}
	messageHMAC := extractMessageHMACFromConfig(config)
	scuttlebutt := formats.NewScuttlebutt(marshaler, messageHMAC)
	v := newFormats(scuttlebutt)
	rawMessageIdentifier := formats.NewRawMessageIdentifier(v)
	messageRepository := bolt.NewMessageRepository(tx, rawMessageIdentifier)
	receiveLogRepository := bolt.NewReceiveLogRepository(tx, messageRepository)
	pubRepository := bolt.NewPubRepository(tx)
	blobRepository := bolt.NewBlobRepository(tx)
	feedRepository := bolt.NewFeedRepository(tx, socialGraphRepository, receiveLogRepository, messageRepository, pubRepository, blobRepository, scuttlebutt)
	currentTimeProvider := adapters.NewCurrentTimeProvider()
	wantListRepository := bolt.NewWantListRepository(tx, currentTimeProvider)
	commandsAdapters := commands.Adapters{
		Feed:        feedRepository,
		SocialGraph: socialGraphRepository,
		WantList:    wantListRepository,
	}
	return commandsAdapters, nil
}

var (
	_wireGraphHopsValue = hops
)

func BuildTxRepositories(tx *bbolt.Tx, public identity.Public, logger logging.Logger, messageHMAC formats.MessageHMAC) (bolt.TxRepositories, error) {
	graphHops := _wireHopsValue2
	socialGraphRepository := bolt.NewSocialGraphRepository(tx, public, graphHops)
	messageContentMappings := transport.DefaultMappings()
	marshaler, err := transport.NewMarshaler(messageContentMappings, logger)
	if err != nil {
		return bolt.TxRepositories{}, err
	}
	scuttlebutt := formats.NewScuttlebutt(marshaler, messageHMAC)
	v := newFormats(scuttlebutt)
	rawMessageIdentifier := formats.NewRawMessageIdentifier(v)
	messageRepository := bolt.NewMessageRepository(tx, rawMessageIdentifier)
	receiveLogRepository := bolt.NewReceiveLogRepository(tx, messageRepository)
	pubRepository := bolt.NewPubRepository(tx)
	blobRepository := bolt.NewBlobRepository(tx)
	feedRepository := bolt.NewFeedRepository(tx, socialGraphRepository, receiveLogRepository, messageRepository, pubRepository, blobRepository, scuttlebutt)
	currentTimeProvider := adapters.NewCurrentTimeProvider()
	wantListRepository := bolt.NewWantListRepository(tx, currentTimeProvider)
	txRepositories := bolt.TxRepositories{
		Feed:       feedRepository,
		Graph:      socialGraphRepository,
		ReceiveLog: receiveLogRepository,
		Message:    messageRepository,
		Blob:       blobRepository,
		WantList:   wantListRepository,
	}
	return txRepositories, nil
}

var (
	_wireHopsValue2 = hops
)

// BuildService creates a new service which uses the provided context as a long-term context used as a base context for
// e.g. established connections.
func BuildService(contextContext context.Context, private identity.Private, config Config) (Service, error) {
	networkKey := extractNetworkKeyFromConfig(config)
	handshaker, err := boxstream.NewHandshaker(private, networkKey)
	if err != nil {
		return Service{}, err
	}
	requestPubSub := pubsub.NewRequestPubSub()
	connectionIdGenerator := rpc.NewConnectionIdGenerator()
	logger := extractLoggerFromConfig(config)
	peerInitializer := transport2.NewPeerInitializer(handshaker, requestPubSub, connectionIdGenerator, logger)
	dialer, err := network.NewDialer(peerInitializer, logger)
	if err != nil {
		return Service{}, err
	}
	db, err := newBolt(config)
	if err != nil {
		return Service{}, err
	}
	public := privateIdentityToPublicIdentity(private)
	adaptersFactory := newAdaptersFactory(config, public)
	transactionProvider := bolt.NewTransactionProvider(db, adaptersFactory)
	messageContentMappings := transport.DefaultMappings()
	marshaler, err := transport.NewMarshaler(messageContentMappings, logger)
	if err != nil {
		return Service{}, err
	}
	redeemInviteHandler := commands.NewRedeemInviteHandler(dialer, transactionProvider, networkKey, private, requestPubSub, marshaler, connectionIdGenerator, logger)
	followHandler := commands.NewFollowHandler(transactionProvider, private, marshaler, logger)
	publishRawHandler := commands.NewPublishRawHandler(transactionProvider, private, logger)
	peerManagerConfig := extractPeerManagerConfigFromConfig(config)
	messageHMAC := extractMessageHMACFromConfig(config)
	txRepositoriesFactory := newTxRepositoriesFactory(public, logger, messageHMAC)
	readContactsRepository := bolt.NewReadContactsRepository(db, txRepositoriesFactory)
	messageBuffer := commands.NewMessageBuffer(transactionProvider, logger)
	manager := replication.NewManager(logger, readContactsRepository, messageBuffer)
	scuttlebutt := formats.NewScuttlebutt(marshaler, messageHMAC)
	v := newFormats(scuttlebutt)
	rawMessageIdentifier := formats.NewRawMessageIdentifier(v)
	rawMessageHandler := commands.NewRawMessageHandler(rawMessageIdentifier, messageBuffer, logger)
	gossipReplicator, err := replication.NewGossipReplicator(manager, rawMessageHandler, logger)
	if err != nil {
		return Service{}, err
	}
	noTxWantListRepository := bolt.NewNoTxWantListRepository(db, txRepositoriesFactory)
	filesystemStorage, err := newFilesystemStorage(logger, config)
	if err != nil {
		return Service{}, err
	}
	blobsGetDownloader := replication2.NewBlobsGetDownloader(filesystemStorage, logger)
	blobDownloadedPubSub := pubsub.NewBlobDownloadedPubSub()
	hasHandler := replication2.NewHasHandler(filesystemStorage, noTxWantListRepository, blobsGetDownloader, blobDownloadedPubSub, logger)
	replicationManager := replication2.NewManager(noTxWantListRepository, filesystemStorage, hasHandler, logger)
	replicator := replication2.NewReplicator(replicationManager)
	peerManager := domain.NewPeerManager(contextContext, peerManagerConfig, gossipReplicator, replicator, dialer, logger)
	connectHandler := commands.NewConnectHandler(peerManager, logger)
	establishNewConnectionsHandler := commands.NewEstablishNewConnectionsHandler(peerManager)
	acceptNewPeerHandler := commands.NewAcceptNewPeerHandler(peerManager)
	processNewLocalDiscoveryHandler := commands.NewProcessNewLocalDiscoveryHandler(peerManager)
	createWantsHandler := commands.NewCreateWantsHandler(replicationManager)
	currentTimeProvider := adapters.NewCurrentTimeProvider()
	downloadBlobHandler := commands.NewDownloadBlobHandler(transactionProvider, currentTimeProvider)
	createBlobHandler := commands.NewCreateBlobHandler(filesystemStorage)
	appCommands := app.Commands{
		RedeemInvite:             redeemInviteHandler,
		Follow:                   followHandler,
		PublishRaw:               publishRawHandler,
		Connect:                  connectHandler,
		EstablishNewConnections:  establishNewConnectionsHandler,
		AcceptNewPeer:            acceptNewPeerHandler,
		ProcessNewLocalDiscovery: processNewLocalDiscoveryHandler,
		CreateWants:              createWantsHandler,
		DownloadBlob:             downloadBlobHandler,
		CreateBlob:               createBlobHandler,
	}
	readFeedRepository := bolt.NewReadFeedRepository(db, txRepositoriesFactory)
	messagePubSub := pubsub.NewMessagePubSub()
	createHistoryStreamHandler := queries.NewCreateHistoryStreamHandler(readFeedRepository, messagePubSub, logger)
	readReceiveLogRepository := bolt.NewReadReceiveLogRepository(db, txRepositoriesFactory)
	receiveLogHandler := queries.NewReceiveLogHandler(readReceiveLogRepository)
	publishedLogHandler, err := queries.NewPublishedLogHandler(readFeedRepository, readReceiveLogRepository, public)
	if err != nil {
		return Service{}, err
	}
	readMessageRepository := bolt.NewReadMessageRepository(db, txRepositoriesFactory)
	statusHandler := queries.NewStatusHandler(readMessageRepository, readFeedRepository, peerManager)
	getBlobHandler, err := queries.NewGetBlobHandler(filesystemStorage)
	if err != nil {
		return Service{}, err
	}
	blobDownloadedEventsHandler := queries.NewBlobDownloadedEventsHandler(blobDownloadedPubSub)
	appQueries := app.Queries{
		CreateHistoryStream:  createHistoryStreamHandler,
		ReceiveLog:           receiveLogHandler,
		PublishedLog:         publishedLogHandler,
		Status:               statusHandler,
		GetBlob:              getBlobHandler,
		BlobDownloadedEvents: blobDownloadedEventsHandler,
	}
	application := app.Application{
		Commands: appCommands,
		Queries:  appQueries,
	}
	listener, err := newListener(contextContext, peerInitializer, application, config, logger)
	if err != nil {
		return Service{}, err
	}
	discoverer, err := local.NewDiscoverer(public, logger)
	if err != nil {
		return Service{}, err
	}
	networkDiscoverer := network2.NewDiscoverer(discoverer, application, logger)
	connectionEstablisher := network2.NewConnectionEstablisher(application, logger)
	handlerBlobsGet := rpc2.NewHandlerBlobsGet(getBlobHandler)
	handlerBlobsCreateWants := rpc2.NewHandlerBlobsCreateWants(createWantsHandler)
	v2 := rpc2.NewMuxHandlers(handlerBlobsGet, handlerBlobsCreateWants)
	handlerCreateHistoryStream := rpc2.NewHandlerCreateHistoryStream(createHistoryStreamHandler, logger)
	v3 := rpc2.NewMuxClosingHandlers(handlerCreateHistoryStream)
	muxMux, err := mux.NewMux(logger, v2, v3)
	if err != nil {
		return Service{}, err
	}
	pubSub := pubsub2.NewPubSub(requestPubSub, muxMux)
	advertiser, err := newAdvertiser(public, config)
	if err != nil {
		return Service{}, err
	}
	service := NewService(application, listener, networkDiscoverer, connectionEstablisher, pubSub, advertiser, messageBuffer, createHistoryStreamHandler)
	return service, nil
}

// wire.go:

type TxTestAdapters struct {
	MessageRepository *bolt.MessageRepository
	FeedRepository    *bolt.FeedRepository
	ReceiveLog        *bolt.ReceiveLogRepository
	WantList          *bolt.WantListRepository

	CurrentTimeProvider *mocks.CurrentTimeProviderMock
}

type TestAdapters struct {
	MessageRepository *bolt.ReadMessageRepository
	FeedRepository    *bolt.ReadFeedRepository
	ReceiveLog        *bolt.ReadReceiveLogRepository
}

type TestQueries struct {
	Queries app.Queries

	FeedRepository       *mocks.FeedRepositoryMock
	MessagePubSub        *mocks.MessagePubSubMock
	MessageRepository    *mocks.MessageRepositoryMock
	PeerManager          *mocks.PeerManagerMock
	BlobStorage          *mocks.BlobStorageMock
	ReceiveLogRepository *mocks.ReceiveLogRepositoryMock

	LocalIdentity identity.Public
}

var replicatorSet = wire.NewSet(replication.NewManager, wire.Bind(new(replication.ReplicationManager), new(*replication.Manager)), replication.NewGossipReplicator, wire.Bind(new(domain.MessageReplicator), new(*replication.GossipReplicator)))

var blobReplicatorSet = wire.NewSet(replication2.NewManager, wire.Bind(new(replication2.ReplicationManager), new(*replication2.Manager)), wire.Bind(new(commands.BlobReplicationManager), new(*replication2.Manager)), replication2.NewReplicator, wire.Bind(new(domain.BlobReplicator), new(*replication2.Replicator)), replication2.NewBlobsGetDownloader, wire.Bind(new(replication2.Downloader), new(*replication2.BlobsGetDownloader)), replication2.NewHasHandler, wire.Bind(new(replication2.HasBlobHandler), new(*replication2.HasHandler)))

var hops = graph.MustNewHops(3)

func newAdvertiser(l identity.Public, config Config) (*local.Advertiser, error) {
	return local.NewAdvertiser(l, config.ListenAddress)
}

func newAdaptersFactory(config Config, local2 identity.Public) bolt.AdaptersFactory {
	return func(tx *bbolt.Tx) (commands.Adapters, error) {
		return BuildTransactableAdapters(tx, local2, config)
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
