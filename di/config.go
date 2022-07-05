package di

import (
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
)

type Config struct {
	// DataDirectory specifies where the primary database and other data
	// will be stored.
	DataDirectory string

	// ListenAddress for the TCP listener in the format accepted by the
	// standard library.
	// Optional, defaults to ":8008".
	ListenAddress string

	// Setting NetworkKey is mainly useful for test networks.
	// Optional, defaults to boxstream.NewDefaultNetworkKey().
	NetworkKey boxstream.NetworkKey

	// Setting MessageHMAC is mainly useful for test networks.
	// Optional, defaults to formats.NewDefaultMessageHMAC().
	MessageHMAC formats.MessageHMAC

	// Logger is the logger used for logging by this library. It is most
	// likely useful to configure at least to log errors.
	// Optional, defaults to logging.NewDevNullLogger().
	Logger logging.Logger

	// PeerManagerConfig specifies the config for the peer manager which is responsible for establishing new
	// connections and managing existing connections.
	PeerManagerConfig domain.PeerManagerConfig
}

func (c *Config) SetDefaults() {
	if c.ListenAddress == "" {
		c.ListenAddress = ":8008"
	}

	if c.NetworkKey.IsZero() {
		c.NetworkKey = boxstream.NewDefaultNetworkKey()
	}

	if c.MessageHMAC.IsZero() {
		c.MessageHMAC = formats.NewDefaultMessageHMAC()
	}

	if c.Logger == nil {
		c.Logger = logging.NewDevNullLogger()
	}
}
