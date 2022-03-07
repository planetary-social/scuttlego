package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/network/rpc/transport"
)

var ErrEndOrErr = errors.New("end or error")

type RawConnection interface {
	Next() (*transport.Message, error)
	Send(msg *transport.Message) error
	Close() error
}

type RequestHandler interface {
	HandleRequest(req *Request, conn *Connection)
}

type Connection struct {
	raw RawConnection

	requestHandler RequestHandler

	incomingRequestNumber int

	outgoingRequestNumber int
	responseStreams       map[int]*ResponseStream
	responseStreamsLock   sync.Mutex

	logger logging.Logger
}

func NewConnection(
	rw io.ReadWriteCloser,
	requestHandler RequestHandler,
	logger logging.Logger,
) (*Connection, error) {
	conn := &Connection{
		raw:             transport.NewRawConnection(rw, logger),
		requestHandler:  requestHandler,
		responseStreams: make(map[int]*ResponseStream),
		logger:          logger.New("connection"),
	}

	go conn.readLoop()

	return conn, nil
}

func (s *Connection) PerformRequest(ctx context.Context, req *Request) (*ResponseStream, error) {
	encodedType, err := s.encodeProcedureType(req.Type())
	if err != nil {
		return nil, errors.Wrap(err, "could not encode the procedure type")
	}

	body := RequestBody{
		Name: req.Name().Components(),
		Type: encodedType,
		Args: req.Arguments(),
	}

	j, err := json.Marshal(body)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal the request body")
	}

	s.logger.WithField("message", string(j)).Debug("sending a raw message")

	s.outgoingRequestNumber++

	flags := transport.MessageHeaderFlags{
		Stream:     req.stream,
		EndOrError: false,
		BodyType:   transport.MessageBodyTypeJSON,
	}

	header, err := transport.NewMessageHeader(
		flags,
		uint32(len(j)),
		int32(s.outgoingRequestNumber),
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a message header")
	}

	msg, err := transport.NewMessage(header, j)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a message")
	}

	stream := s.newResponseStream(&msg)

	if err := s.raw.Send(&msg); err != nil {
		return nil, errors.Wrap(err, "could not send a message")
	}

	return stream, nil
}

func (s *Connection) newResponseStream(msg *transport.Message) *ResponseStream {
	s.responseStreamsLock.Lock()
	defer s.responseStreamsLock.Unlock()

	number := msg.Header.RequestNumber()

	if _, ok := s.responseStreams[number]; ok {
		panic("response stream with this number already exists")
	}

	rs := NewResponseStream(s, number)
	s.responseStreams[number] = rs
	return rs
}

func (s *Connection) closeAllResponseStreams() {
	s.responseStreamsLock.Lock()
	defer s.responseStreamsLock.Unlock()

	for number := range s.responseStreams {
		s.closeResponseStreamWithoutLock(number)
	}
}

func (s *Connection) closeResponseStream(number int) {
	s.responseStreamsLock.Lock()
	defer s.responseStreamsLock.Unlock()

	s.closeResponseStreamWithoutLock(number)
}

func (s *Connection) closeResponseStreamWithoutLock(number int) {
	stream, ok := s.responseStreams[number]
	if !ok {
		return
	}

	delete(s.responseStreams, number)
	close(stream.ch)
}

func (s *Connection) Close() error {
	return s.raw.Close()
}

func (c *Connection) readLoop() {
	defer c.raw.Close()
	defer c.closeAllResponseStreams()

	for {
		if err := c.read(); err != nil {
			c.logger.WithError(err).Debug("shutting down the read loop")
			return
		}
	}
}

func (c *Connection) read() error {
	msg, err := c.raw.Next()
	if err != nil {
		return errors.Wrap(err, "failed to read the next message")
	}

	c.logger.
		WithField("number", msg.Header.RequestNumber()).
		WithField("body", string(msg.Body)).
		Debug("received a new message")

	if msg.Header.IsRequest() {
		return c.handleIncomingRequest(msg)
	} else {
		return c.handleIncomingResponse(msg)
	}
}

const (
	transportStringForProcedureTypeSource = "source"
	transportStringForProcedureTypeDuplex = "duplex"
	transportStringForProcedureTypeAsync  = "async"
)

func (s *Connection) decodeProcedureType(str string) ProcedureType {
	switch str {
	case transportStringForProcedureTypeSource:
		return ProcedureTypeSource
	case transportStringForProcedureTypeDuplex:
		return ProcedureTypeDuplex
	case transportStringForProcedureTypeAsync:
		return ProcedureTypeAsync
	default:
		return ProcedureTypeUnknown
	}
}

func (s *Connection) encodeProcedureType(t ProcedureType) (string, error) {
	switch t {
	case ProcedureTypeSource:
		return transportStringForProcedureTypeSource, nil
	case ProcedureTypeDuplex:
		return transportStringForProcedureTypeDuplex, nil
	case ProcedureTypeAsync:
		return transportStringForProcedureTypeAsync, nil
	default:
		return "", fmt.Errorf("unknown procedure type %T", t)
	}
}

func (s *Connection) handleIncomingRequest(msg *transport.Message) error {
	s.incomingRequestNumber += 1
	if got, expected := msg.Header.RequestNumber(), s.incomingRequestNumber; got != expected {
		return fmt.Errorf("invalid request number (got: %d, expected: %d)", got, expected)
	}

	var requestBody RequestBody
	if err := json.Unmarshal(msg.Body, &requestBody); err != nil {
		return errors.Wrap(err, "could not unmarshal the request body")
	}

	procedureName, err := NewProcedureName(requestBody.Name)
	if err != nil {
		return errors.Wrap(err, "could not create a procedure name")
	}

	procedureType := s.decodeProcedureType(requestBody.Type)

	req, err := NewRequest(
		procedureName,
		procedureType,
		msg.Header.Flags().Stream,
		requestBody.Args,
	)
	if err != nil {
		return errors.Wrap(err, "could not create a request")
	}

	go s.requestHandler.HandleRequest(req, s)

	return nil
}

func (s *Connection) handleIncomingResponse(msg *transport.Message) error {
	s.responseStreamsLock.Lock()
	defer s.responseStreamsLock.Unlock()

	rs, ok := s.responseStreams[-msg.Header.RequestNumber()]
	if !ok {
		return nil
	}

	if msg.Header.Flags().EndOrError {
		rs.ch <- ResponseWithError{
			Value: NewResponse(msg.Body),
			Err:   ErrEndOrErr,
		}
		s.closeResponseStreamWithoutLock(rs.number)
	} else {
		rs.ch <- ResponseWithError{
			Value: NewResponse(msg.Body),
			Err:   nil,
		}
	}

	return nil
}

type ResponseWriter struct {
	req  *Request
	conn *Connection
}

func NewResponseWriter(req *Request, conn *Connection) ResponseWriter {
	return ResponseWriter{
		req:  req,
		conn: conn,
	}
}

func (rw ResponseWriter) OpenResponseStream(bodyType transport.MessageBodyType) io.WriteCloser {
	panic("not implemented")
}

type ResponseWithError struct {
	Value *Response
	Err   error
}

type ResponseStream struct {
	ch     chan ResponseWithError
	conn   *Connection
	number int
}

func NewResponseStream(conn *Connection, number int) *ResponseStream {
	return &ResponseStream{
		ch:     make(chan ResponseWithError),
		number: number,
		conn:   conn,
	}
}

func (rs ResponseStream) Channel() <-chan ResponseWithError {
	return rs.ch
}

func (rs ResponseStream) Close() error {
	rs.conn.closeResponseStream(rs.number)
	return nil
}
