package replication

import (
	"context"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/messages"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/transport"
)

type WantsProcess struct {
	lock     sync.Mutex
	incoming []IncomingStream

	wantListStorage WantListStorage
	blobStorage     BlobStorage
	downloader      Downloader
	logger          logging.Logger
}

func NewWantsProcess(
	wantListStorage WantListStorage,
	blobStorage BlobStorage,
	downloader Downloader,
	logger logging.Logger,
) *WantsProcess {
	return &WantsProcess{
		wantListStorage: wantListStorage,
		blobStorage:     blobStorage,
		downloader:      downloader,
		logger:          logger.New("wants_process"),
	}
}

func (p *WantsProcess) AddIncoming(stream IncomingStream) {
	p.logger.Debug("adding incoming")

	p.lock.Lock()
	defer p.lock.Unlock()

	p.incoming = append(p.incoming, stream)
	go p.incomingLoop(stream)
}

func (p *WantsProcess) AddOutgoing(ctx context.Context, ch <-chan messages.BlobWithSizeOrWantDistance, peer transport.Peer) {
	p.logger.Debug("adding outgoing")
	go p.outgoingLoop(ctx, ch, peer)
}

func (p *WantsProcess) incomingLoop(stream IncomingStream) {
	defer close(stream.ch)
	// todo cleanup?

	for {
		wl, err := p.wantListStorage.GetWantList()
		if err != nil {
			p.logger.WithError(err).Error("could not get the want list")
			continue
		}

		for _, v := range wl.List() {
			v, err := messages.NewBlobWithWantDistance(v.Id, v.Distance)
			if err != nil {
				p.logger.WithError(err).Error("could not create a blob with want distance")
				continue
			}

			p.logger.WithField("blob", v.Id()).Debug("sending wants")

			select {
			case stream.ch <- v:
				continue
			case <-stream.ctx.Done():
				return
			}
		}

		select {
		case <-stream.ctx.Done():
			return
		case <-time.After(10 * time.Second): // todo change
			continue
		}
	}
}

func (p *WantsProcess) outgoingLoop(ctx context.Context, ch <-chan messages.BlobWithSizeOrWantDistance, peer transport.Peer) {
	for hasOrWant := range ch {
		logger := p.logger.WithField("blob", hasOrWant.Id().String())

		if size, ok := hasOrWant.SizeOrWantDistance().Size(); ok {
			logger.WithField("size", size.InBytes()).Debug("received has")
			go p.downloader.OnHasReceived(ctx, peer, hasOrWant.Id(), size)
			continue
		}

		if distance, ok := hasOrWant.SizeOrWantDistance().WantDistance(); ok {
			logger.WithField("distance", distance.Int()).Debug("received want")

			if err := p.onReceiveWant(hasOrWant.Id()); err != nil {
				logger.WithError(err).Error("error processing a want")
			}

			continue
		}

		panic("logic error")
	}

	// todo channel closed, cleanup?
}

func (p *WantsProcess) onReceiveWant(id refs.Blob) error {
	size, err := p.blobStorage.Size(id)
	if err != nil {
		if errors.Is(err, ErrBlobNotFound) {
			p.logger.WithField("blob", id).Debug("we don't have this blob")
			return nil
		}
		return errors.Wrap(err, "could not get blob size")
	}

	has, err := messages.NewBlobWithSize(id, size)
	if err != nil {
		return errors.Wrap(err, "could not create a has")
	}

	p.sendToAll(has)
	return nil
}

func (p *WantsProcess) sendToAll(wantOrHas messages.BlobWithSizeOrWantDistance) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.logger.
		WithField("v", wantOrHas).
		WithField("num_incoming", len(p.incoming)).
		Debug("sending want or has")

	for _, incoming := range p.incoming {
		incoming.ch <- wantOrHas // todo what if this is slow
	}
}

type IncomingStream struct {
	ctx context.Context
	ch  chan<- messages.BlobWithSizeOrWantDistance
}

func NewIncomingStream(ctx context.Context, ch chan<- messages.BlobWithSizeOrWantDistance) IncomingStream {
	return IncomingStream{
		ctx: ctx,
		ch:  ch,
	}
}
