package rpc

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc/transport"
)

type MessageSender interface {
	Send(msg *transport.Message) error
}

// ResponseStreams is used for handling streams initiated by us (for which
// incoming messages have negative request numbers).
type ResponseStreams struct {
	closed      bool
	streams     map[int]*ResponseStream
	streamsLock sync.Mutex

	outgoingRequestNumber uint32

	raw    MessageSender
	logger logging.Logger
}

func NewResponseStreams(raw MessageSender, logger logging.Logger) *ResponseStreams {
	return &ResponseStreams{
		raw:     raw,
		streams: make(map[int]*ResponseStream),
		logger:  logger.New("response_streams"),
	}
}

func (s *ResponseStreams) Open(ctx context.Context, req *Request) (*ResponseStream, error) {
	msg, err := s.marshalRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal a request")
	}

	s.streamsLock.Lock()
	defer s.streamsLock.Unlock()

	if s.closed {
		return nil, errors.New("response streams were closed")
	}

	requestNumber := msg.Header.RequestNumber()

	if _, ok := s.streams[requestNumber]; ok {
		return nil, errors.New("response stream with this number already exists")
	}

	rs := NewResponseStream(ctx, requestNumber)
	s.streams[requestNumber] = rs

	go s.waitAndCloseResponseStream(rs)

	if err := s.raw.Send(msg); err != nil {
		return nil, errors.Wrap(err, "could not send a message")
	}

	return rs, nil
}

// HandleIncomingResponse processes an incoming response. Returning an error
// from this function shuts down the entire connection.
func (s *ResponseStreams) HandleIncomingResponse(msg *transport.Message) error {
	if msg.Header.IsRequest() {
		return errors.New("passed a request")
	}

	s.streamsLock.Lock()
	defer s.streamsLock.Unlock()

	rs, ok := s.streams[-msg.Header.RequestNumber()]
	if !ok {
		return nil
	}

	var err error
	if msg.Header.Flags().EndOrError() {
		err = ErrEndOrErr
		defer rs.cancel()
	}

	select {
	case rs.ch <- ResponseWithError{Value: NewResponse(msg.Body), Err: err}:
	case <-rs.ctx.Done():
	}

	return nil
}

func (s *ResponseStreams) marshalRequest(req *Request) (*transport.Message, error) {
	requestNumber := s.newOutgoingRequestNumber()
	return marshalRequest(req, requestNumber)
}

func (s *ResponseStreams) newOutgoingRequestNumber() uint32 {
	return atomic.AddUint32(&s.outgoingRequestNumber, 1)
}

func (s *ResponseStreams) Close() {
	s.streamsLock.Lock()
	defer s.streamsLock.Unlock()

	s.closed = true

	for _, rs := range s.streams {
		rs.cancel()
	}
}

func (s *ResponseStreams) waitAndCloseResponseStream(rs *ResponseStream) {
	<-rs.ctx.Done()

	s.streamsLock.Lock()
	defer s.streamsLock.Unlock()

	go func() {
		if err := sendCloseStream(s.raw, rs.number, nil); err != nil {
			s.logger.WithError(err).Debug("failed to close the stream")
		}
	}()

	delete(s.streams, rs.number)
	close(rs.ch)
}

type ResponseStream struct {
	number int
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan ResponseWithError
}

func NewResponseStream(ctx context.Context, number int) *ResponseStream {
	ctx, cancel := context.WithCancel(ctx)

	rs := &ResponseStream{
		number: number,
		ctx:    ctx,
		cancel: cancel,
		ch:     make(chan ResponseWithError),
	}

	return rs
}

func (rs ResponseStream) Channel() <-chan ResponseWithError {
	return rs.ch
}

type ResponseWithError struct {
	Value *Response
	Err   error
}
