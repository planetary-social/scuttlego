package replication

import (
	"context"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type WantsProcess struct {
	incomingLock sync.Mutex
	incoming     []incomingStream

	remoteWantsLock sync.Mutex
	remoteWants     map[string]struct{}

	wantedBlobsProvider             WantedBlobsProvider
	blobsThatShouldBePushedProvider BlobsThatShouldBePushedProvider
	blobStorage                     BlobSizeRepository
	hasHandler                      HasBlobHandler
	logger                          logging.Logger
}

func NewWantsProcess(
	wantedBlobsProvider WantedBlobsProvider,
	blobsThatShouldBePushedProvider BlobsThatShouldBePushedProvider,
	blobStorage BlobSizeRepository,
	hasHandler HasBlobHandler,
	logger logging.Logger,
) *WantsProcess {
	return &WantsProcess{
		remoteWants: make(map[string]struct{}),

		wantedBlobsProvider:             wantedBlobsProvider,
		blobsThatShouldBePushedProvider: blobsThatShouldBePushedProvider,
		blobStorage:                     blobStorage,
		hasHandler:                      hasHandler,
		logger:                          logger.New("wants_process"),
	}
}

func (p *WantsProcess) AddIncoming(ctx context.Context, ch chan<- messages.BlobWithSizeOrWantDistance) {
	p.logger.Trace("adding incoming")

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
	p.logger.Trace("adding outgoing")
	go p.outgoingLoop(ctx, ch, peer)
}

func (p *WantsProcess) incomingLoop(ctx context.Context, ch chan<- messages.BlobWithSizeOrWantDistance) {
	if err := p.respondToPreviousWants(); err != nil {
		p.logger.WithError(err).Error("failed to respond to previous wants")
	}

	for {
		if err := p.sendWantList(ctx, ch); err != nil {
			if !errors.Is(err, context.Canceled) {
				p.logger.WithError(err).Error("error sending want list")
			}
		}

		select {
		case <-time.After(10 * time.Second):
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (p *WantsProcess) sendWantList(ctx context.Context, ch chan<- messages.BlobWithSizeOrWantDistance) error {
	wantedBlobs, err := p.wantedBlobsProvider.GetWantedBlobs()
	if err != nil {
		return errors.Wrap(err, "could not get the want list")
	}

	blobsThatShouldBePushed, err := p.blobsThatShouldBePushedProvider.GetBlobsThatShouldBePushed()
	if err != nil {
		return errors.Wrap(err, "could not get the want list")
	}

	wantList, err := BuildWantList(wantedBlobs, blobsThatShouldBePushed)
	if err != nil {
		return errors.Wrap(err, "could not build the want list")
	}

	p.logger.WithField("want_list_size", wantList.Len()).Trace("built the want list")

	for _, wantedBlob := range wantList.List() {
		v, err := messages.NewBlobWithWantDistance(wantedBlob.Id, wantedBlob.Distance)
		if err != nil {
			return errors.Wrap(err, "could not create a blob with want distance")
		}

		p.logger.
			WithField("blob", wantedBlob.Id.String()).
			WithField("distance", wantedBlob.Distance.Int()).
			Trace("sending wants")

		select {
		case ch <- v:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
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

func BuildWantList(wantedBlobs []refs.Blob, blobsToPush []refs.Blob) (blobs.WantList, error) {
	wantListMap := make(map[string]blobs.WantedBlob)

	for _, blob := range wantedBlobs {
		wantListMap[blob.String()] = blobs.WantedBlob{
			Id:       blob,
			Distance: blobs.NewWantDistanceLocal(),
		}
	}

	for _, blob := range blobsToPush {
		if _, has := wantListMap[blob.String()]; !has {
			wantListMap[blob.String()] = blobs.WantedBlob{
				Id:       blob,
				Distance: blobs.NewWantDistanceLocal(),
			}
		}
	}

	var wantListSlice []blobs.WantedBlob
	for _, v := range wantListMap {
		wantListSlice = append(wantListSlice, v)
	}

	wantList, err := blobs.NewWantList(wantListSlice)
	if err != nil {
		return blobs.WantList{}, errors.Wrap(err, "error creating the want list")
	}

	return wantList, nil
}
