package replication

import (
	"context"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type WantsProcess struct {
	incomingLock sync.Mutex
	incoming     []incomingStream

	remoteWantsLock sync.Mutex
	remoteWants     map[string]struct{}

	wantListStorage WantListStorage
	blobStorage     BlobSizeRepository
	hasHandler      HasBlobHandler
	logger          logging.Logger
}

func NewWantsProcess(
	wantListStorage WantListStorage,
	blobStorage BlobSizeRepository,
	hasHandler HasBlobHandler,
	logger logging.Logger,
) *WantsProcess {
	return &WantsProcess{
		remoteWants: make(map[string]struct{}),

		wantListStorage: wantListStorage,
		blobStorage:     blobStorage,
		hasHandler:      hasHandler,
		logger:          logger.New("wants_process"),
	}
}

func (p *WantsProcess) AddIncoming(ctx context.Context, ch chan<- messages.BlobWithSizeOrWantDistance) {
	p.logger.Debug("adding incoming")

	p.incomingLock.Lock()
	defer p.incomingLock.Unlock()

	stream := newIncomingStream(ctx, ch)
	p.incoming = append(p.incoming, stream)

	go func() {
		defer close(ch)
		defer func() {
			if err := p.cleanupIncomingStream(stream); err != nil {
				p.logger.WithError(err).Error("could not clean up a stream")
			}
		}()
		p.incomingLoop(stream.ctx, stream.ch)
	}()
}

func (p *WantsProcess) AddOutgoing(ctx context.Context, ch <-chan messages.BlobWithSizeOrWantDistance, peer transport.Peer) {
	p.logger.Debug("adding outgoing")
	go p.outgoingLoop(ctx, ch, peer)
}

func (p *WantsProcess) incomingLoop(ctx context.Context, ch chan<- messages.BlobWithSizeOrWantDistance) {
	if err := p.respondToPreviousWants(); err != nil {
		p.logger.WithError(err).Error("failed to respond to previous wants")
	}

	for {
		wl, err := p.wantListStorage.List()
		if err != nil {
			p.logger.WithError(err).Error("could not get the want list")
			continue
		}

		p.logger.WithField("want_list_size", wl.Len()).Trace("want list loaded from storage")

		for _, v := range wl.List() {
			v, err := messages.NewBlobWithWantDistance(v.Id, v.Distance)
			if err != nil {
				p.logger.WithError(err).Error("could not create a blob with want distance")
				continue
			}

			p.logger.WithField("blob", v.Id()).Debug("sending wants")

			select {
			case ch <- v:
				continue
			case <-ctx.Done():
				return
			}
		}

		select {
		case <-time.After(10 * time.Second): // todo change
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (p *WantsProcess) cleanupIncomingStream(stream incomingStream) error {
	p.incomingLock.Lock()
	defer p.incomingLock.Unlock()

	for i := range p.incoming {
		if p.incoming[i] == stream {
			p.incoming = append(p.incoming[:i], p.incoming[i+1:]...)
			return nil
		}
	}
	return errors.New("incoming stream not found")
}

func (p *WantsProcess) outgoingLoop(ctx context.Context, ch <-chan messages.BlobWithSizeOrWantDistance, peer transport.Peer) {
	for hasOrWant := range ch {
		logger := p.logger.WithField("blob", hasOrWant.Id().String())

		if size, ok := hasOrWant.SizeOrWantDistance().Size(); ok {
			logger.WithField("size", size.InBytes()).Debug("received has")

			go p.hasHandler.OnHasReceived(ctx, peer, hasOrWant.Id(), size)
			continue
		}

		if distance, ok := hasOrWant.SizeOrWantDistance().WantDistance(); ok {
			logger.WithField("distance", distance.Int()).Trace("received want")

			if err := p.onReceiveWant(hasOrWant.Id()); err != nil {
				logger.WithError(err).Error("error processing a want")
			}
			continue
		}

		panic("logic error")
	}
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
			p.logger.WithField("blob", id).Trace("we don't have this blob")
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
