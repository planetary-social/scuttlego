package network

import (
	"context"

	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/network/local"
)

type ProcessNewLocalDiscoveryCommandHandler interface {
	Handle(ctx context.Context, cmd commands.ProcessNewLocalDiscovery) error
}

// Discoverer receives local UDP announcements from other Secure Scuttlebutt
// clients and passes them to the ProcessNewLocalDiscovery command.
type Discoverer struct {
	discoverer *local.Discoverer
	handler    ProcessNewLocalDiscoveryCommandHandler
	logger     logging.Logger
}

func NewDiscoverer(
	discoverer *local.Discoverer,
	handler ProcessNewLocalDiscoveryCommandHandler,
	logger logging.Logger,
) *Discoverer {
	return &Discoverer{
		discoverer: discoverer,
		handler:    handler,
		logger:     logger.New("discoverer"),
	}
}

// Run receives local announcements and passes them to the command until the
// context is closed.
func (d Discoverer) Run(ctx context.Context) error {
	for v := range d.discoverer.Run(ctx) {
		go d.handleNewDiscovery(ctx, v)
	}

	return nil
}

func (d Discoverer) handleNewDiscovery(ctx context.Context, v local.IdentityWithAddress) {
	if err := d.handler.Handle(
		ctx,
		commands.ProcessNewLocalDiscovery{
			Remote:  v.Remote,
			Address: v.Address,
		},
	); err != nil {
		d.logger.WithError(err).Error("failed to handle a discovered peer")
	}
}
