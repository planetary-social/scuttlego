package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service"
	"github.com/planetary-social/scuttlego/service/domain"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
)

var extractFromConfigSet = wire.NewSet(
	extractNetworkKeyFromConfig,
	extractMessageHMACFromConfig,
	extractLoggingSystemFromConfig,
	extractPeerManagerConfigFromConfig,
	extractHopsFromConfig,
)

func extractNetworkKeyFromConfig(config service.Config) boxstream.NetworkKey {
	return config.NetworkKey
}

func extractMessageHMACFromConfig(config service.Config) formats.MessageHMAC {
	return config.MessageHMAC
}

func extractLoggingSystemFromConfig(config service.Config) logging.LoggingSystem {
	return config.LoggingSystem
}

func extractPeerManagerConfigFromConfig(config service.Config) domain.PeerManagerConfig {
	return config.PeerManagerConfig
}

func extractHopsFromConfig(config service.Config) graph.Hops {
	return *config.Hops
}
