package replication

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/blobs"
	"github.com/planetary-social/go-ssb/service/domain/messages"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/transport"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
)

type WantListStorage interface {
	GetWantList() (blobs.WantList, error)
}

type Downloader interface {
	OnHasReceived(ctx context.Context, peer transport.Peer, blob refs.Blob, size blobs.Size)
}

type Manager struct {
	storage    WantListStorage
	downloader Downloader
	logger     logging.Logger
}

func NewManager(storage WantListStorage, downloader Downloader, logger logging.Logger) *Manager {
	return &Manager{
		storage:    storage,
		downloader: downloader,
		logger:     logger.New("replication_manager"),
	}
}

func (r *Manager) HandleIncomingCreateWantsRequest(ctx context.Context) (<-chan messages.BlobWithSizeOrWantDistance, error) {
	connectionId, ok := rpc.GetConnectionIdFromContext(ctx)
	if !ok {
		return nil, errors.New("connection id not found in context")
	}
	r.logger.WithField("connectionId", connectionId).Debug("incoming create wants")

	ch := make(chan messages.BlobWithSizeOrWantDistance)
	go r.sendWantListPeriodically(ctx, ch)
	return ch, nil
}

func (r *Manager) HandleOutgoingCreateWantsRequest(ctx context.Context, ch <-chan messages.BlobWithSizeOrWantDistance, peer transport.Peer) error {
	connectionId, ok := rpc.GetConnectionIdFromContext(ctx)
	if !ok {
		return errors.New("connection id not found in context")
	}
	r.logger.WithField("connectionId", connectionId).Debug("outgoing create wants")

	go r.handleOutgoing(ctx, connectionId, ch, peer)
	return nil
}

func (r *Manager) handleOutgoing(ctx context.Context, id rpc.ConnectionId, ch <-chan messages.BlobWithSizeOrWantDistance, peer transport.Peer) {
	for blobWithSizeOrWantDistance := range ch {
		logger := r.logger.WithField("connection_id", id).WithField("blob", blobWithSizeOrWantDistance.Id().String())

		if size, ok := blobWithSizeOrWantDistance.SizeOrWantDistance().Size(); ok {
			logger.WithField("size", size.InBytes()).Debug("got size")
			go r.downloader.OnHasReceived(ctx, peer, blobWithSizeOrWantDistance.Id(), size)
			continue
		}

		if distance, ok := blobWithSizeOrWantDistance.SizeOrWantDistance().WantDistance(); ok {
			// peer wants a blob
			// todo tell it that we have it if we have it
			logger.WithField("distance", distance.Int()).Debug("got distance")
			continue
		}

		panic("logic error")
	}

	// todo channel closed
}

func (r *Manager) sendWantListPeriodically(ctx context.Context, ch chan<- messages.BlobWithSizeOrWantDistance) {
	defer close(ch)
	defer r.logger.Debug("terminating sending want list")

	for {
		wl, err := r.storage.GetWantList()
		if err != nil {
			r.logger.WithError(err).Error("could not get the want list")
			continue
		}

		for _, v := range wl.List() {
			v, err := messages.NewBlobWithWantDistance(v.Id, v.Distance)
			if err != nil {
				r.logger.WithError(err).Error("could not create a blob with want distance")
				continue
			}

			r.logger.WithField("blob", v.Id()).Debug("sending wants")

			select {
			case ch <- v:
				continue
			case <-ctx.Done():
				return
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second): // todo change
			continue
		}
	}
}
