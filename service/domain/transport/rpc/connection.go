package rpc

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc/transport"
)

var ErrEndOrErr = errors.New("end or error")

type RawConnection interface {
	Next() (*transport.Message, error)
	Send(msg *transport.Message) error
	Close() error
}

type Connection struct {
	raw RawConnection

	responseStreams *ResponseStreams
	requestStreams  *RequestStreams

	doneCh chan struct{}

	logger logging.Logger
}

func NewConnection(
	raw RawConnection,
	handler RequestHandler,
	logger logging.Logger,
) (*Connection, error) {
	conn := &Connection{
		raw:             raw,
		responseStreams: NewResponseStreams(raw, logger),
		requestStreams:  NewRequestStreams(raw, handler, logger),
		logger:          logger.New("connection"),
		doneCh:          make(chan struct{}),
	}

	go conn.readLoop()

	return conn, nil
}

func (s *Connection) PerformRequest(ctx context.Context, req *Request) (*ResponseStream, error) {
	stream, err := s.responseStreams.Open(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open a response stream")
	}

	return stream, nil
}

func (s *Connection) Done() <-chan struct{} {
	return s.doneCh
}

func (s *Connection) Close() error {
	return s.raw.Close()
}

func (c *Connection) readLoop() {
	defer c.raw.Close()
	defer c.responseStreams.Close()
	defer c.requestStreams.Close()
	defer close(c.doneCh)

	for {
		if err := c.read(); err != nil {
			c.logger.WithError(err).Debug("read loop shutting down")
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
		return c.requestStreams.HandleIncomingRequest(msg)
	} else {
		return c.responseStreams.HandleIncomingResponse(msg)
	}
}
