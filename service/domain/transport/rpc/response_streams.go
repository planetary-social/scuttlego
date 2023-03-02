package rpc

import (
	"bytes"
	"context"
	"net"
	"sync"
	"sync/atomic"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
)

type MessageSender interface {
	Send(msg *transport.Message) error
}

// ResponseStream represents a stream that we initiated.
type ResponseStream interface {
	WriteMessage(body []byte, bodyType transport.MessageBodyType) error
	Channel() <-chan ResponseWithError
	Ctx() context.Context
}

var (
	// ErrRemoteEnd signals that the remote closed the stream but didn't signal
	// that an error occurred.
	ErrRemoteEnd = errors.New("remote end")
)

type RemoteError struct {
	response []byte
}

func NewRemoteError(response []byte) error {
	return &RemoteError{response: response}
}

func (e RemoteError) Error() string {
	return "remote returned an error"
}

func (e RemoteError) Response() []byte {
	return e.response
}

func (e RemoteError) As(target interface{}) bool {
	if v, ok := target.(*RemoteError); ok {
		*v = e
		return true
	}
	return false
}

func (e RemoteError) Is(target error) bool {
	_, ok1 := target.(*RemoteError)
	_, ok2 := target.(RemoteError)
	return ok1 || ok2
}

// todo private fields and constructor
type ResponseWithError struct {
	// Value is only set if Err is nil.
	Value *Response

	// If Err is not nil then it may be of ErrRemoteEnd, RemoteError or a
	// different error.
	Err error
}

// ResponseStreams is used for handling streams initiated by us (for which
// incoming messages have negative request numbers).
type ResponseStreams struct {
	closed      bool
	streams     map[int]*responseStream
	streamsLock sync.Mutex

	outgoingRequestNumber uint32

	raw    MessageSender
	logger logging.Logger
}

func NewResponseStreams(raw MessageSender, logger logging.Logger) *ResponseStreams {
	return &ResponseStreams{
		raw:     raw,
		streams: make(map[int]*responseStream),
		logger:  logger.New("response_streams"),
	}
}

func (s *ResponseStreams) Open(ctx context.Context, req *Request) (ResponseStream, error) {
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

	rs, err := newResponseStream(ctx, requestNumber, req.Type(), s.raw)
	if err != nil {
		return nil, errors.Wrap(err, "error creating a response stream")
	}

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

	if msg.Header.Flags().EndOrError() {
		defer rs.cancel()
		rs.handleRemoteErr(msg.Body)
	} else {
		rs.handleRemoteResponse(NewResponse(msg.Body))
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

func (s *ResponseStreams) waitAndCloseResponseStream(rs *responseStream) {
	<-rs.ctx.Done()

	s.streamsLock.Lock()
	defer s.streamsLock.Unlock()

	if rs.typ != ProcedureTypeAsync {
		go func() {
			if err := sendCloseStream(s.raw, rs.number, nil); err != nil {
				if !errors.Is(err, net.ErrClosed) {
					s.logger.WithError(err).Debug("failed to close the stream")
				}
			}
		}()
	}

	delete(s.streams, rs.number)
	close(rs.ch)
}

type responseStream struct {
	number int
	typ    ProcedureType
	ctx    context.Context
	cancel context.CancelFunc
	raw    MessageSender

	ch chan ResponseWithError
}

func newResponseStream(ctx context.Context, number int, typ ProcedureType, raw MessageSender) (*responseStream, error) {
	if number <= 0 {
		return nil, errors.New("number must be positive")
	}

	ctx = logging.AddToLoggingContext(ctx, logging.StreamIdContextLabel, number)
	ctx, cancel := context.WithCancel(ctx)

	return &responseStream{
		number: number,
		typ:    typ,
		ctx:    ctx,
		cancel: cancel,
		raw:    raw,
		ch:     make(chan ResponseWithError),
	}, nil
}

func (rs responseStream) WriteMessage(body []byte, bodyType transport.MessageBodyType) error {
	if rs.typ != ProcedureTypeDuplex {
		return errors.New("not a duplex stream")
	}

	if bodyType.IsZero() {
		return errors.New("zero value of body type")
	}

	select {
	case <-rs.ctx.Done():
		return rs.ctx.Err()
	default:
	}

	// todo is the stream flag correct?
	flags, err := transport.NewMessageHeaderFlags(true, false, bodyType)
	if err != nil {
		return errors.Wrap(err, "could not create message header flags")
	}

	header, err := transport.NewMessageHeader(flags, uint32(len(body)), int32(rs.number))
	if err != nil {
		return errors.Wrap(err, "could not create a message header")
	}

	msg, err := transport.NewMessage(header, body)
	if err != nil {
		return errors.Wrap(err, "could not create a message")
	}

	if err := rs.raw.Send(&msg); err != nil {
		return errors.Wrap(err, "could not send a message")
	}

	return nil
}

func (rs responseStream) Ctx() context.Context {
	return rs.ctx
}

func (rs responseStream) Channel() <-chan ResponseWithError {
	return rs.ch
}

func (rs responseStream) handleRemoteErr(body []byte) {
	select {
	case rs.ch <- ResponseWithError{Err: rs.guessError(body)}:
	case <-rs.ctx.Done():
	}
}

func (rs responseStream) handleRemoteResponse(resp *Response) {
	select {
	case rs.ch <- ResponseWithError{Value: resp}:
	case <-rs.ctx.Done():
	}
}

func (rs responseStream) guessError(body []byte) error {
	if bytes.Equal([]byte("true"), body) {
		return ErrRemoteEnd
	}
	return NewRemoteError(body)
}
