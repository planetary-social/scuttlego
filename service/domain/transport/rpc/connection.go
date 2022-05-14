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
	ctx    context.Context
	cancel context.CancelFunc

	raw RawConnection

	responseStreams *ResponseStreams
	requestStreams  *RequestStreams

	logger logging.Logger
}

// NewConnection is the only way of creating a new Connection, zero value is invalid. Terminating the provided context
// is equivalent to calling Close. The provided context is used as a base context for the contexts passed to the
// request handler. Connection takes over managing RawConnection which must not be used further.
func NewConnection(
	ctx context.Context,
	raw RawConnection,
	handler RequestHandler,
	logger logging.Logger,
) (*Connection, error) {
	ctx, cancel := context.WithCancel(ctx)

	conn := &Connection{
		ctx:             ctx,
		cancel:          cancel,
		raw:             raw,
		responseStreams: NewResponseStreams(raw, logger),
		requestStreams:  NewRequestStreams(ctx, raw, handler, logger),
		logger:          logger.New("connection"),
	}

	go func() {
		<-ctx.Done()
		defer conn.raw.Close()
		defer conn.responseStreams.Close()
	}()

	go conn.readLoop()

	return conn, nil
}

func (c *Connection) PerformRequest(ctx context.Context, req *Request) (*ResponseStream, error) {
	stream, err := c.responseStreams.Open(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open a response stream")
	}

	return stream, nil
}

func (c *Connection) Done() <-chan struct{} {
	return c.ctx.Done()
}

// Close always returns nil. In theory shutting down a Secure Scuttlebutt RPC connection can result in an error as
// a goodbye message for the entire connection has to be sent successfully to the other side but those errors are
// not made available as it is unclear what to do with them.
func (c *Connection) Close() error {
	c.cancel()
	return nil
}

func (c *Connection) readLoop() {
	defer c.cancel()

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
		return c.requestStreams.HandleIncomingRequest(c.ctx, msg)
	} else {
		return c.responseStreams.HandleIncomingResponse(msg)
	}
}
