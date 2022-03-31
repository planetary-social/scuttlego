package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc/transport"
)

type RequestHandler interface {
	HandleRequest(ctx context.Context, req *Request, rw *ResponseWriter)
}

type responseWriters map[int]*ResponseWriter

type RequestStreams struct {
	closed      bool
	writers     responseWriters
	writersLock sync.Mutex

	incomingRequestNumber int

	handler RequestHandler

	raw    MessageSender
	logger logging.Logger
}

func NewRequestStreams(raw MessageSender, handler RequestHandler, logger logging.Logger) *RequestStreams {
	return &RequestStreams{
		raw:     raw,
		writers: make(responseWriters),
		handler: handler,
		logger:  logger.New("response_streams"),
	}
}

func (r *RequestStreams) HandleIncomingRequest(msg *transport.Message) error {
	if !msg.Header.IsRequest() {
		return errors.New("passed a response")
	}

	r.writersLock.Lock()
	defer r.writersLock.Unlock()

	if msg.Header.Flags().EndOrError {
		return r.terminateWriter(msg)
	} else {
		return r.openWriter(msg)
	}
}

func (r *RequestStreams) terminateWriter(msg *transport.Message) error {
	requestNumber := msg.Header.RequestNumber()

	rw, ok := r.writers[requestNumber]
	if !ok {
		return nil
	}

	rw.cancel()
	return nil
}

func (r *RequestStreams) openWriter(msg *transport.Message) error {
	r.incomingRequestNumber += 1
	if got, expected := msg.Header.RequestNumber(), r.incomingRequestNumber; got != expected {
		return fmt.Errorf("invalid request number (got: %d, expected: %d)", got, expected)
	}

	req, err := r.unmarshalRequest(msg)
	if err != nil {
		return errors.Wrap(err, "unmarshal request failed")
	}

	requestNumber := msg.Header.RequestNumber()

	_, ok := r.writers[requestNumber]
	if ok {
		return errors.New("duplicate request number")
	}

	rw := NewResponseWriter(requestNumber, r.raw)
	r.writers[requestNumber] = rw
	go r.waitAndCleanupWriter(rw)

	go r.handler.HandleRequest(context.TODO(), req, rw) // todo add a real context associated with the connection

	return nil
}

func (s *RequestStreams) unmarshalRequest(msg *transport.Message) (*Request, error) {
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
		msg.Header.Flags().Stream,
		requestBody.Args,
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a request")
	}

	return req, err
}

func (s *RequestStreams) Close() {
	s.writersLock.Lock()
	defer s.writersLock.Unlock()

	s.closed = true

	for _, rs := range s.writers {
		rs.cancel()
	}
}

func (s *RequestStreams) waitAndCleanupWriter(rw *ResponseWriter) {
	<-rw.ctx.Done()

	s.writersLock.Lock()
	defer s.writersLock.Unlock()

	delete(s.writers, rw.requestNumber)
}

type ResponseWriter struct {
	requestNumber int

	ctx    context.Context
	cancel context.CancelFunc
	raw    MessageSender
}

func NewResponseWriter(number int, raw MessageSender) *ResponseWriter {
	ctx, cancel := context.WithCancel(context.TODO())

	rs := &ResponseWriter{
		requestNumber: number,
		raw:           raw,

		ctx:    ctx,
		cancel: cancel,
	}

	return rs
}

func (rs ResponseWriter) WriteMessage(body []byte) error {
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

func (rs ResponseWriter) CloseWithError(err error) error {
	rs.cancel()
	return rs.sendCloseStream(err)
}

func (rs *ResponseWriter) sendCloseStream(err error) error {
	return sendCloseStream(rs.raw, -rs.requestNumber, err)
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
