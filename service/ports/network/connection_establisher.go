package network

import (
	"context"
	"time"

	"github.com/planetary-social/scuttlego/logging"
)

type EstablishNewConnectionsCommandHandler interface {
	Handle(ctx context.Context) error
}

// ConnectionEstablisher periodically triggers the EstablishNewConnections
// application command.
type ConnectionEstablisher struct {
	establishConnectionsEvery time.Duration
	handler                   EstablishNewConnectionsCommandHandler
	logger                    logging.Logger
}

func NewConnectionEstablisher(
	handler EstablishNewConnectionsCommandHandler,
	logger logging.Logger,
) *ConnectionEstablisher {
	return &ConnectionEstablisher{
		establishConnectionsEvery: 15 * time.Second,
		handler:                   handler,
		logger:                    logger.New("connection_establisher"),
	}
}

// Run periodically triggers the command until the context is closed.
func (d ConnectionEstablisher) Run(ctx context.Context) error {
	for {
		if err := d.handler.Handle(ctx); err != nil {
			d.logger.WithError(err).Debug("failed to establish new connections")
		}

		select {
		case <-time.After(d.establishConnectionsEvery):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
