package rpc

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc/transport"
)

type ResponseWriter interface {
	WriteMessage(body []byte) error
	CloseWithError(err error) error
}

type RequestHandler interface {
	// HandleRequest must eventually call ResponseWriter.CloseWithError to avoid memory leaks. HandleRequest may block
	// indefinitely.
	HandleRequest(ctx context.Context, rw ResponseWriter, req *Request)
}

// RequestStreams is used for handling incoming data (with positive request numbers).
type RequestStreams struct {
	closed bool

	writers       map[int]*responseWriter
	closedWriters map[int]struct{}
	writersLock   sync.Mutex // guards writers and closedWriters

	handler RequestHandler

	raw    MessageSender
	logger logging.Logger
}

func NewRequestStreams(raw MessageSender, handler RequestHandler, logger logging.Logger) *RequestStreams {
	return &RequestStreams{
		raw:           raw,
		writers:       make(map[int]*responseWriter),
		closedWriters: make(map[int]struct{}),
		handler:       handler,
		logger:        logger.New("response_streams"),
	}
}

func (s *RequestStreams) HandleIncomingRequest(msg *transport.Message) error {
	s.writersLock.Lock()
	defer s.writersLock.Unlock()

	if !msg.Header.IsRequest() {
		return errors.New("passed a response")
	}

	if s.closed {
		return errors.New("streams closed")
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

	if msg.Header.Flags().EndOrError {
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

	rw.cancel()
}

func (s *RequestStreams) openNewWriter(msg *transport.Message) error {
	requestNumber := msg.Header.RequestNumber()

	req, err := unmarshalRequest(msg)
	if err != nil {
		return errors.Wrap(err, "unmarshal request failed")
	}

	rw := newResponseWriter(requestNumber, req.Type(), s.raw)
	s.writers[requestNumber] = rw
	go s.waitAndCleanupWriter(rw)

	go s.handler.HandleRequest(context.TODO(), rw, req) // todo add a real context associated with the connection

	return nil
}

func (s *RequestStreams) waitAndCleanupWriter(rw *responseWriter) {
	<-rw.ctx.Done()

	s.writersLock.Lock()
	defer s.writersLock.Unlock()

	delete(s.writers, rw.requestNumber)
	s.closedWriters[rw.requestNumber] = struct{}{}
}

func (s *RequestStreams) Close() {
	s.writersLock.Lock()
	defer s.writersLock.Unlock()

	s.closed = true

	for _, rs := range s.writers {
		rs.cancel()
	}
}

type responseWriter struct {
	requestNumber int
	typ           ProcedureType

	ctx    context.Context
	cancel context.CancelFunc
	raw    MessageSender
}

func newResponseWriter(number int, typ ProcedureType, raw MessageSender) *responseWriter {
	ctx, cancel := context.WithCancel(context.TODO())

	rs := &responseWriter{
		requestNumber: number,
		typ:           typ,

		ctx:    ctx,
		cancel: cancel,
		raw:    raw,
	}

	return rs
}

func (rs responseWriter) WriteMessage(body []byte) error {
	// todo do this correctly? are the flags correct?
	flags := transport.MessageHeaderFlags{
		Stream:     true,
		EndOrError: false,
		BodyType:   transport.MessageBodyTypeJSON,
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

func (rs responseWriter) CloseWithError(err error) error {
	rs.cancel()
	return rs.sendCloseStream(err)
}

func (rs *responseWriter) sendCloseStream(err error) error {
	return sendCloseStream(rs.raw, -rs.requestNumber, err)
}

func (rs responseWriter) handleNewMessage(msg *transport.Message) error {
	if rs.typ != ProcedureTypeDuplex {
		return errors.New("illegal duplicate request number")
	}

	// todo pass msg to the handler
	return nil
}

func sendCloseStream(raw MessageSender, number int, err error) error {
	// todo do this correctly? are the flags correct?
	flags := transport.MessageHeaderFlags{
		Stream:     true,
		EndOrError: true,
		BodyType:   transport.MessageBodyTypeJSON,
	}

	var content []byte
	if err == nil {
		content = []byte("true")
	} else {
		var mErr error
		content, mErr = json.Marshal(struct {
			Error string `json:"error"`
		}{err.Error()})
		if mErr != nil {
			panic(mErr)
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
