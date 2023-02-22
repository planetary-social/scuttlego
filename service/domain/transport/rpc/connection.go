package rpc

import (
	"context"
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
)

type RawConnection interface {
	Next() (*transport.Message, error)
	Send(msg *transport.Message) error
	Close() error
}

type Connection struct {
	id                   ConnectionId
	wasInitiatedByRemote bool

	raw RawConnection

	responseStreams *ResponseStreams
	requestStreams  *RequestStreams

	logger logging.Logger
}

// NewConnection is the only way of creating a new Connection, zero value is invalid. Terminating the provided context
// is equivalent to calling Close. The provided context is used as a base context for the contexts passed to the
// request handler. Connection takes over managing RawConnection which must not be used further.
func NewConnection(
	id ConnectionId,
	wasInitiatedByRemote bool,
	raw RawConnection,
	handler RequestHandler,
	logger logging.Logger,
) (*Connection, error) {
	conn := &Connection{
		wasInitiatedByRemote: wasInitiatedByRemote,
		raw:                  raw,
		responseStreams:      NewResponseStreams(raw, logger),
		requestStreams:       NewRequestStreams(raw, handler, logger),
		logger:               logger.New("connection"),
		id:                   id,
	}

	return conn, nil
}

func (c *Connection) Loop(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer c.raw.Close()
	defer c.responseStreams.Close()

	go c.requestStreams.cleanupLoop(ctx)

	for {
		if err := c.read(ctx); err != nil {
			return errors.Wrap(err, "read returned an error")
		}
	}
}

func (c *Connection) PerformRequest(ctx context.Context, req *Request) (ResponseStream, error) {
	stream, err := c.responseStreams.Open(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open a response stream")
	}

	return stream, nil
}

// Close always returns nil. In theory shutting down a Secure Scuttlebutt RPC connection can result in an error as
// a goodbye message for the entire connection has to be sent successfully to the other side but those errors are
// not made available as it is unclear what to do with them.
func (c *Connection) Close() error {
	return c.raw.Close()
}

func (c *Connection) WasInitiatedByRemote() bool {
	return c.wasInitiatedByRemote
}

func (c *Connection) String() string {
	return fmt.Sprintf("<id=%s initiatedByRemote=%t>", c.id, c.wasInitiatedByRemote)
}

func (c *Connection) read(ctx context.Context) error {
	msg, err := c.raw.Next()
	if err != nil {
		return errors.Wrap(err, "failed to read the next message")
	}

	if msg.Header.IsRequest() {
		return c.requestStreams.HandleIncomingRequest(ctx, msg)
	} else {
		return c.responseStreams.HandleIncomingResponse(msg)
	}
}
