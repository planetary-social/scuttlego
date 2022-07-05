package network

import (
	"context"
	"time"

	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/app"
)

// ConnectionEstablisher periodically triggers the EstablishNewConnections
// application command.
type ConnectionEstablisher struct {
	establishConnectionsEvery time.Duration
	app                       app.Application
	logger                    logging.Logger
}

func NewConnectionEstablisher(
	app app.Application,
	logger logging.Logger,
) *ConnectionEstablisher {
	return &ConnectionEstablisher{
		establishConnectionsEvery: 15 * time.Second,
		app:                       app,
		logger:                    logger.New("connection_establisher"),
	}
}

// Run periodically triggers the command until the context is closed.
func (d ConnectionEstablisher) Run(ctx context.Context) error {
	for {
		if err := d.app.Commands.EstablishNewConnections.Handle(); err != nil {
			d.logger.WithError(err).Debug("failed to establish new connections")
		}

		select {
		case <-time.After(d.establishConnectionsEvery):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
