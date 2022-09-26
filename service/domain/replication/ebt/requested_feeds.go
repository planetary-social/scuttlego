package ebt

import (
	"context"
	"sync"

	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type RequestedFeeds struct {
	messageStreamer MessageStreamer
	stream          Stream

	activeStreams     map[string]context.CancelFunc
	activeStreamsLock sync.Mutex
}

func NewRequestedFeeds(messageStreamer MessageStreamer, stream Stream) *RequestedFeeds {
	return &RequestedFeeds{
		messageStreamer: messageStreamer,
		stream:          stream,
		activeStreams:   make(map[string]context.CancelFunc),
	}
}

func (r *RequestedFeeds) Request(ctx context.Context, ref refs.Feed, seq *message.Sequence) {
	r.activeStreamsLock.Lock()
	defer r.activeStreamsLock.Unlock()

	r.cancelIfExists(ref)

	ctx, cancel := context.WithCancel(ctx)
	r.messageStreamer.Handle(ctx, ref, seq, NewStreamMessageWriter(r.stream))
	r.activeStreams[ref.String()] = cancel
}

func (r *RequestedFeeds) Cancel(ref refs.Feed) {
	r.activeStreamsLock.Lock()
	defer r.activeStreamsLock.Unlock()

	r.cancelIfExists(ref)
}

func (r *RequestedFeeds) cancelIfExists(ref refs.Feed) {
	if cancelExisting, ok := r.activeStreams[ref.String()]; ok {
		cancelExisting()
	}
	delete(r.activeStreams, ref.String())
}
