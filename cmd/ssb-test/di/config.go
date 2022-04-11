package di

import (
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/feeds/formats"
	"github.com/planetary-social/go-ssb/service/domain/transport/boxstream"
)

type Config struct {
	LoggingLevel  logging.Level
	DataDirectory string
	ListenAddress string

	NetworkKey  boxstream.NetworkKey
	MessageHMAC formats.MessageHMAC
}

func (c *Config) SetDefaults() {
	if c.NetworkKey.IsZero() {
		c.NetworkKey = boxstream.NewDefaultNetworkKey()
	}

	if c.MessageHMAC.IsZero() {
		c.MessageHMAC = formats.NewDefaultMessageHMAC()
	}
}
