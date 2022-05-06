package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain"
	"github.com/planetary-social/go-ssb/service/domain/feeds/formats"
	"github.com/planetary-social/go-ssb/service/domain/transport/boxstream"
)

//nolint:deadcode,varcheck
var extractFromConfigSet = wire.NewSet(
	extractNetworkKeyFromConfig,
	extractMessageHMACFromConfig,
	extractLoggerFromConfig,
	extractPeerManagerConfigFromConfig,
)

func extractNetworkKeyFromConfig(config Config) boxstream.NetworkKey {
	return config.NetworkKey
}

func extractMessageHMACFromConfig(config Config) formats.MessageHMAC {
	return config.MessageHMAC
}

func extractLoggerFromConfig(config Config) logging.Logger {
	return config.Logger
}

func extractPeerManagerConfigFromConfig(config Config) domain.PeerManagerConfig {
	return config.PeerManagerConfig
}
