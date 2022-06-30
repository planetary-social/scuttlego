package replication

import (
	"context"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/blobs"
	"github.com/planetary-social/go-ssb/service/domain/messages"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/transport"
)

type BlobSizeRepository interface {
	// Size returns the size of the blob. If the blob is not found it returns
	// ErrBlobNotFound.
	Size(id refs.Blob) (blobs.Size, error)
}

type WantsProcess struct {
	incomingLock sync.Mutex
	incoming     []incomingStream

	remoteWantsLock sync.Mutex
	remoteWants     map[string]struct{}

	wantListStorage WantListStorage
	blobStorage     BlobSizeRepository
	downloader      Downloader
	logger          logging.Logger
}

func NewWantsProcess(
	wantListStorage WantListStorage,
	blobStorage BlobSizeRepository,
	downloader Downloader,
	logger logging.Logger,
) *WantsProcess {
	return &WantsProcess{
		remoteWants: make(map[string]struct{}),

		wantListStorage: wantListStorage,
		blobStorage:     blobStorage,
		downloader:      downloader,
		logger:          logger.New("wants_process"),
	}
}

func (p *WantsProcess) AddIncoming(ctx context.Context, ch chan<- messages.BlobWithSizeOrWantDistance) {
	p.logger.Debug("adding incoming")

	p.incomingLock.Lock()
	defer p.incomingLock.Unlock()

	stream := newIncomingStream(ctx, ch)
	p.incoming = append(p.incoming, stream)
	go p.incomingLoop(stream)
}

func (p *WantsProcess) AddOutgoing(ctx context.Context, ch <-chan messages.BlobWithSizeOrWantDistance, peer transport.Peer) {
	p.logger.Debug("adding outgoing")
	go p.outgoingLoop(ctx, ch, peer)
}

func (p *WantsProcess) incomingLoop(stream incomingStream) {
	defer close(stream.ch)
	// todo cleanup?

	if err := p.respondToPreviousWants(); err != nil {
		p.logger.WithError(err).Error("failed to respond to previous wants")
	}

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
	p.addRemoteWants(id)
	if err := p.respondToWant(id); err != nil {
		return errors.Wrap(err, "could not respond to want")
	}
	return nil
}

func (p *WantsProcess) addRemoteWants(id refs.Blob) {
	p.remoteWantsLock.Lock()
	defer p.remoteWantsLock.Unlock()
	p.remoteWants[id.String()] = struct{}{}
}

func (p *WantsProcess) respondToPreviousWants() error {
	p.remoteWantsLock.Lock()
	defer p.remoteWantsLock.Unlock()

	for refString := range p.remoteWants {
		if err := p.respondToWant(refs.MustNewBlob(refString)); err != nil {
			return errors.Wrap(err, "failed to respond to a want")
		}
	}

	return nil
}

func (p *WantsProcess) respondToWant(id refs.Blob) error {
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

	p.sendToAllIncomingStreams(has)
	return nil
}

func (p *WantsProcess) sendToAllIncomingStreams(wantOrHas messages.BlobWithSizeOrWantDistance) {
	p.incomingLock.Lock()
	defer p.incomingLock.Unlock()

	p.logger.
		WithField("v", wantOrHas).
		WithField("num_incoming", len(p.incoming)).
		Debug("sending want or has")

	for _, incoming := range p.incoming {
		select {
		case incoming.ch <- wantOrHas: // todo what if this is slow
		case <-incoming.ctx.Done():
			continue
		}
	}
}

type incomingStream struct {
	ctx context.Context
	ch  chan<- messages.BlobWithSizeOrWantDistance
}

func newIncomingStream(ctx context.Context, ch chan<- messages.BlobWithSizeOrWantDistance) incomingStream {
	return incomingStream{
		ctx: ctx,
		ch:  ch,
	}
}
