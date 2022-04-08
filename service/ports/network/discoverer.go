package network

import (
	"context"

	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/app"
	"github.com/planetary-social/go-ssb/service/app/commands"
	"github.com/planetary-social/go-ssb/service/domain/network/local"
)

type Discoverer struct {
	discoverer *local.Discoverer
	app        app.Application
	logger     logging.Logger
}

func NewDiscoverer(discoverer *local.Discoverer, app app.Application, logger logging.Logger) *Discoverer {
	return &Discoverer{
		discoverer: discoverer,
		app:        app,
		logger:     logger.New("discoverer"),
	}
}

func (d Discoverer) Run(ctx context.Context) error {
	for v := range d.discoverer.Run(ctx) {
		if err := d.handleNewDiscovery(v); err != nil {
			d.logger.WithError(err).Error("failed to handle a discovered peer")
		}
	}

	return nil
}

func (d Discoverer) handleNewDiscovery(v local.IdentityWithAddress) error {
	return d.app.Commands.ProcessNewLocalDiscovery.Handle(
		commands.ProcessNewLocalDiscovery{
			Remote:  v.Remote,
			Address: v.Address,
		},
	)
}
