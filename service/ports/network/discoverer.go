package network

import (
	"context"

	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/app"
	"github.com/planetary-social/go-ssb/service/app/commands"
	"github.com/planetary-social/go-ssb/service/domain/network/local"
)

// Discoverer receives local UDP announcements from other Secure Scuttlebutt
// clients and passes them to the ProcessNewLocalDiscovery command.
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

// Run receives local announcements and passes them to the command until the
// context is closed.
func (d Discoverer) Run(ctx context.Context) error {
	for v := range d.discoverer.Run(ctx) {
		go d.handleNewDiscovery(v)
	}

	return nil
}

func (d Discoverer) handleNewDiscovery(v local.IdentityWithAddress) {
	if err := d.app.Commands.ProcessNewLocalDiscovery.Handle(
		commands.ProcessNewLocalDiscovery{
			Remote:  v.Remote,
			Address: v.Address,
		},
	); err != nil {
		d.logger.WithError(err).Error("failed to handle a discovered peer")
	}
}
