package di

import (
	"context"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/adapters/badger/notx"
	"github.com/planetary-social/scuttlego/service/app"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/network/local"
	networkport "github.com/planetary-social/scuttlego/service/ports/network"
	pubsubport "github.com/planetary-social/scuttlego/service/ports/pubsub"
)

type Service struct {
	App app.Application

	listener                     *networkport.Listener
	discoverer                   *networkport.Discoverer
	connectionEstablisher        *networkport.ConnectionEstablisher
	requestSubscriber            *pubsubport.RequestSubscriber
	roomAttendantEventSubscriber *pubsubport.RoomAttendantEventSubscriber
	newPeerSubscriber            *pubsubport.NewPeerSubscriber
	advertiser                   *local.Advertiser
	messageBuffer                *commands.MessageBuffer
	createHistoryStreamHandler   *queries.CreateHistoryStreamHandler
	badgerGarbageCollector       *badger.GarbageCollector
	feedWantListRepository       *notx.NoTxFeedWantListRepository
	blobWantListRepository       *notx.NoTxBlobWantListRepository
}

func NewService(
	app app.Application,
	listener *networkport.Listener,
	discoverer *networkport.Discoverer,
	connectionEstablisher *networkport.ConnectionEstablisher,
	requestSubscriber *pubsubport.RequestSubscriber,
	roomAttendantEventSubscriber *pubsubport.RoomAttendantEventSubscriber,
	newPeerSubscriber *pubsubport.NewPeerSubscriber,
	advertiser *local.Advertiser,
	messageBuffer *commands.MessageBuffer,
	createHistoryStreamHandler *queries.CreateHistoryStreamHandler,
	badgerGarbageCollector *badger.GarbageCollector,
	feedWantListRepository *notx.NoTxFeedWantListRepository,
	blobWantListRepository *notx.NoTxBlobWantListRepository,
) Service {
	return Service{
		App: app,

		listener:                     listener,
		discoverer:                   discoverer,
		connectionEstablisher:        connectionEstablisher,
		requestSubscriber:            requestSubscriber,
		roomAttendantEventSubscriber: roomAttendantEventSubscriber,
		newPeerSubscriber:            newPeerSubscriber,
		advertiser:                   advertiser,
		messageBuffer:                messageBuffer,
		createHistoryStreamHandler:   createHistoryStreamHandler,
		badgerGarbageCollector:       badgerGarbageCollector,
		feedWantListRepository:       feedWantListRepository,
		blobWantListRepository:       blobWantListRepository,
	}
}

func (s Service) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error)
	runners := 0

	runners++
	go func() {
		errCh <- s.listener.ListenAndServe(ctx)
	}()

	runners++
	go func() {
		errCh <- s.requestSubscriber.Run(ctx)
	}()

	runners++
	go func() {
		errCh <- s.roomAttendantEventSubscriber.Run(ctx)
	}()

	runners++
	go func() {
		errCh <- s.newPeerSubscriber.Run(ctx)
	}()

	runners++
	go func() {
		errCh <- s.advertiser.Run(ctx)
	}()

	runners++
	go func() {
		errCh <- s.discoverer.Run(ctx)
	}()

	runners++
	go func() {
		errCh <- s.connectionEstablisher.Run(ctx)
	}()

	runners++
	go func() {
		errCh <- s.messageBuffer.Run(ctx)
	}()

	runners++
	go func() {
		errCh <- s.createHistoryStreamHandler.Run(ctx)
	}()

	runners++
	go func() {
		errCh <- s.badgerGarbageCollector.Run(ctx)
	}()

	runners++
	go func() {
		errCh <- s.feedWantListRepository.CleanupLoop(ctx)
	}()

	runners++
	go func() {
		errCh <- s.blobWantListRepository.CleanupLoop(ctx)
	}()

	var err error
	for i := 0; i < runners; i++ {
		err = multierror.Append(err, errors.Wrap(<-errCh, "error returned by runner"))
		cancel()
	}

	return err
}
