package commands

import (
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/network"
)

type ProcessNewLocalDiscovery struct {
	Remote  identity.Public
	Address network.Address
}

type ProcessNewLocalDiscoveryHandler struct {
	logger logging.Logger
}

func NewProcessNewLocalDiscoveryHandler(logger logging.Logger) *ProcessNewLocalDiscoveryHandler {
	return &ProcessNewLocalDiscoveryHandler{
		logger: logger,
	}
}

func (h *ProcessNewLocalDiscoveryHandler) Handle(cmd ProcessNewLocalDiscovery) error {
	h.logger.WithField("address", cmd.Address).Debug("new local peer discovered")
	return nil
}
