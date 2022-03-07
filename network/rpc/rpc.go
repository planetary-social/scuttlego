package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

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
	responseStreams       *ResponseStreams

	logger logging.Logger
}

func NewConnection(
	rw io.ReadWriteCloser,
	requestHandler RequestHandler,
	logger logging.Logger,
) (*Connection, error) {
	raw := transport.NewRawConnection(rw, logger)
	conn := &Connection{
		raw:             raw,
		requestHandler:  requestHandler,
		responseStreams: NewResponseStreams(raw, logger),
		logger:          logger.New("connection"),
	}

	go conn.readLoop()

	return conn, nil
}

func (s *Connection) PerformRequest(ctx context.Context, req *Request) (*ResponseStream, error) {
	encodedType, err := encodeProcedureType(req.Type())
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

	stream, err := s.responseStreams.Open(ctx, msg.Header.RequestNumber())
	if err != nil {
		return nil, errors.Wrap(err, "failed to open a response stream")
	}

	if err := s.raw.Send(&msg); err != nil {
		return nil, errors.Wrap(err, "could not send a message")
	}

	return stream, nil
}

func (s *Connection) Close() error {
	return s.raw.Close()
}

func (c *Connection) readLoop() {
	defer c.raw.Close()
	defer c.responseStreams.Close()

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

	if msg.Header.IsRequest() {
		return c.handleIncomingRequest(msg)
	} else {
		return c.responseStreams.HandleIncomingResponse(msg)
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

	procedureType := decodeProcedureType(requestBody.Type)

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

const (
	transportStringForProcedureTypeSource = "source"
	transportStringForProcedureTypeDuplex = "duplex"
	transportStringForProcedureTypeAsync  = "async"
)

func decodeProcedureType(str string) ProcedureType {
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

func encodeProcedureType(t ProcedureType) (string, error) {
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
