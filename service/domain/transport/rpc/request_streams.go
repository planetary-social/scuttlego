package rpc

import (
	"context"
	"encoding/json"
	"net"
	"sync"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc/transport"
)

type ResponseWriter interface {
	// WriteMessage sends a message over the underlying stream.
	WriteMessage(body []byte) error

	// CloseWithError terminates the underlying stream. Error is sent to the other party. Error can be nil.
	CloseWithError(err error) error
}

type RequestHandler interface {
	// HandleRequest should respond to the provided request using the response writer. Implementations must eventually
	// call ResponseWriter.CloseWithError. HandleRequest may block as it is executed in a goroutine.
	HandleRequest(ctx context.Context, rw ResponseWriter, req *Request)
}

// RequestStreams is used for handling incoming data (with positive request numbers).
type RequestStreams struct {
	ctx context.Context

	writers       map[int]*requestStream
	closedWriters map[int]struct{}
	writersLock   sync.Mutex // guards writers and closedWriters

	handler RequestHandler

	raw    MessageSender
	logger logging.Logger
}

// NewRequestStreams create new RequestStreams which use the provided context as a base context for contexts provided
// to the RequestHandler.
func NewRequestStreams(ctx context.Context, raw MessageSender, handler RequestHandler, logger logging.Logger) *RequestStreams {
	return &RequestStreams{
		ctx:           ctx,
		raw:           raw,
		writers:       make(map[int]*requestStream),
		closedWriters: make(map[int]struct{}),
		handler:       handler,
		logger:        logger.New("response_streams"),
	}
}

// HandleIncomingRequest processes an incoming request. Returning an error from this function shuts down the entire
// connection.
func (s *RequestStreams) HandleIncomingRequest(msg *transport.Message) error {
	s.writersLock.Lock()
	defer s.writersLock.Unlock()

	if !msg.Header.IsRequest() {
		return errors.New("passed a response")
	}

	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
	}

	requestNumber := msg.Header.RequestNumber()

	// Unfortunately we are not able to distinguish "start of stream" messages from "part of stream" messages in the
	// case of duplex connections. This is a problem which exists at the design level of the Secure Scuttlebutt's RPC
	// protocol. Both the initial message in a duplex connection and follow-up messages:
	// - have the stream flag set
	// - have the same request number
	// It is therefore impossible to distinguish them without really weird payload-level heuristics or broken
	// bookkeeping like the one applied here which may eventually cause the program to run out of memory.
	//
	// Real scenario:
	// If we close a duplex connection and then a new message with the same request number comes in we have to remember
	// that we should discard it instead of opening a new stream based on this message. We have no other means to
	// realise that this message is not a new request.
	//
	// This could be partially mitigated if we assumed that new request numbers can only be larger than previous
	// request numbers but protocol docs make no such assumptions. Technically they don't also mention that request
	// numbers can't be reused but there is simply no way to implement the RPC protocol without making that assumption.
	if _, ok := s.closedWriters[requestNumber]; ok {
		return nil
	}

	if msg.Header.Flags().EndOrError() {
		s.tryTerminateWriter(requestNumber)
		return nil
	}

	existingWriter, ok := s.writers[requestNumber]
	if ok {
		return existingWriter.handleNewMessage(msg)
	}

	return s.openNewWriter(msg)
}

func (s *RequestStreams) tryTerminateWriter(requestNumber int) {
	rw, ok := s.writers[requestNumber]
	if !ok {
		return
	}

	rw.terminatedByRemote()
}

func (s *RequestStreams) openNewWriter(msg *transport.Message) error {
	requestNumber := msg.Header.RequestNumber()

	req, err := unmarshalRequest(msg)
	if err != nil {
		s.logger.WithError(err).Debug("ignoring malformed request")
		if sendCloseStreamErr := sendCloseStream(s.raw, -msg.Header.RequestNumber(), errors.New("received malformed request")); sendCloseStreamErr != nil {
			if !errors.Is(sendCloseStreamErr, net.ErrClosed) {
				s.logger.WithError(sendCloseStreamErr).Debug("error sending close stream message")
			}
		}
		return nil
	}

	rw := newRequestStream(s.ctx, requestNumber, req.Type(), s.raw)
	s.writers[requestNumber] = rw
	go s.waitAndCleanupWriter(rw)

	go s.handler.HandleRequest(rw.ctx, rw, req)

	return nil
}

func (s *RequestStreams) waitAndCleanupWriter(rw *requestStream) {
	<-rw.ctx.Done()

	s.writersLock.Lock()
	defer s.writersLock.Unlock()

	delete(s.writers, rw.requestNumber)
	s.closedWriters[rw.requestNumber] = struct{}{}
}

type requestStream struct {
	requestNumber int
	typ           ProcedureType

	closed      bool
	closedMutex sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	raw    MessageSender
}

func newRequestStream(ctx context.Context, number int, typ ProcedureType, raw MessageSender) *requestStream {
	ctx, cancel := context.WithCancel(ctx)

	rs := &requestStream{
		requestNumber: number,
		typ:           typ,

		ctx:    ctx,
		cancel: cancel,
		raw:    raw,
	}

	return rs
}

func (rs *requestStream) WriteMessage(body []byte) error {
	if err := rs.ensureNotClosed(); err != nil {
		return err
	}

	// todo do this correctly? are the flags correct?
	flags, err := transport.NewMessageHeaderFlags(true, false, transport.MessageBodyTypeJSON)
	if err != nil {
		return errors.Wrap(err, "could not create message header flags")
	}

	header, err := transport.NewMessageHeader(flags, uint32(len(body)), int32(-rs.requestNumber))
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

func (rs *requestStream) CloseWithError(err error) error {
	rs.closedMutex.Lock()
	defer rs.closedMutex.Unlock()

	if rs.closed {
		return errors.New("already closed")
	}

	rs.cancel()
	rs.closed = true
	return sendCloseStream(rs.raw, -rs.requestNumber, err)
}

func (rs *requestStream) ensureNotClosed() error {
	rs.closedMutex.Lock()
	defer rs.closedMutex.Unlock()

	if rs.closed {
		return errors.New("already closed")
	}

	return nil
}

func (rs *requestStream) terminatedByRemote() {
	rs.cancel()
}

func (rs *requestStream) handleNewMessage(msg *transport.Message) error {
	if rs.typ != ProcedureTypeDuplex {
		return errors.New("illegal duplicate request number")
	}

	// todo pass msg to the handler

	return nil
}

func unmarshalRequest(msg *transport.Message) (*Request, error) {
	var requestBody RequestBody
	if err := json.Unmarshal(msg.Body, &requestBody); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal the request body")
	}

	procedureName, err := NewProcedureName(requestBody.Name)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a procedure name")
	}

	procedureType := decodeProcedureType(requestBody.Type)

	req, err := NewRequest(
		procedureName,
		procedureType,
		requestBody.Args,
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a request")
	}

	return req, err
}

// sendCloseStream closes the specified stream. Number is put directly into the
// request number field of the sent message. If you are closing a stream that
// you initiated then number will be a positive value, if you are closing a
// stream that a peer initiated then number will be a negative value.
func sendCloseStream(raw MessageSender, number int, errToSent error) error {
	// todo do this correctly? are the flags correct?
	flags, err := transport.NewMessageHeaderFlags(true, true, transport.MessageBodyTypeJSON)
	if err != nil {
		return errors.Wrap(err, "could not create message header flags")
	}

	var content []byte
	if errToSent == nil {
		content = []byte("true") // todo why true, is there any reason for this? do we have to send something specific? is this documented?
	} else {
		var mErr error
		content, mErr = json.Marshal(struct {
			Error string `json:"error"`
		}{errToSent.Error()})
		if mErr != nil {
			panic(mErr) // tests would have caught this eg. TestPrematureTerminationByRemote
		}
	}

	header, err := transport.NewMessageHeader(flags, uint32(len(content)), int32(number))
	if err != nil {
		return errors.Wrap(err, "could not create a message header")
	}

	msg, err := transport.NewMessage(header, content)
	if err != nil {
		return errors.Wrap(err, "could not create a message")
	}

	if err := raw.Send(&msg); err != nil {
		return errors.Wrap(err, "could not send a message")
	}

	return nil
}
